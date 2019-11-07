package ara

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"
)

type simpleAddr struct {
	addr    string
	network string
}

func (addr simpleAddr) String() string {
	return addr.addr
}

func (addr simpleAddr) Network() string {
	return addr.network
}

type timeoutErr struct{}

func (timeoutErr) Error() string {
	return "timeout"
}

func (timeoutErr) Timeout() bool {
	return true
}

func (timeoutErr) Temporary() bool {
	return false
}

var errMissingAddress = errors.New("missing address")

// Dialer is a partial replacement for net.Dialer but it still
// uses net.Dialer internally.
//
// Unlike net.Dialer, it accepts an actual customizable Resolver.
//
// A Dialer contains options for connecting to an address.
//
// The zero value for each field is equivalent to dialing
// without that option.
type Dialer struct {
	// Timeout is the maximum amount of time a dial will wait for
	// a connect to complete. If Deadline is also set, it may fail
	// earlier.
	//
	// The default is no timeout.
	//
	// When using TCP and dialing a host name with multiple IP
	// addresses, the timeout may be divided between them.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	Timeout time.Duration

	// Deadline is the absolute point in time after which dials
	// will fail. If Timeout is set, it may fail earlier.
	// Zero means no deadline, or dependent on the operating system
	// as with the Timeout option.
	Deadline time.Time

	// LocalAddr is the local address to use when dialing an
	// address. The address must be of a compatible type for the
	// network being dialed.
	// If nil, a local address is automatically chosen.
	LocalAddr net.Addr

	// FallbackDelay specifies the length of time to wait before
	// spawning a RFC 6555 Fast Fallback connection. That is, this
	// is the amount of time to wait for IPv6 to succeed before
	// assuming that IPv6 is misconfigured and falling back to
	// IPv4.
	//
	// If zero, a default delay of 300ms is used.
	// A negative value disables Fast Fallback support.
	FallbackDelay time.Duration

	// KeepAlive specifies the interval between keep-alive
	// probes for an active network connection.
	// If zero, keep-alive probes are sent with a default value
	// (currently 15 seconds), if supported by the protocol and operating
	// system. Network protocols or operating systems that do
	// not support keep-alives ignore this field.
	// If negative, keep-alive probes are disabled.
	KeepAlive time.Duration

	// Resolver optionally specifies an alternate resolver to use.
	Resolver Resolver

	// If Control is not nil, it is called after creating the network
	// connection but before actually dialing.
	//
	// Network and address parameters passed to Control method are not
	// necessarily the ones passed to Dial. For example, passing "tcp" to Dial
	// will cause the Control function to be called with "tcp4" or "tcp6".
	Control func(network, address string, c syscall.RawConn) error

	// Underlying dialer
	d *net.Dialer
}

// DialContext connects to the address on the named network using
// the provided context.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
//
// When using TCP, and the host in the address parameter resolves to multiple
// network addresses, any dial timeout (from d.Timeout or ctx) is spread
// over each consecutive dial, such that each is given an appropriate
// fraction of the time to connect.
// For example, if a host has 4 IP addresses and the timeout is 1 minute,
// the connect to each single address will be given 15 seconds to complete
// before trying the next one.
//
// See func net.Dial for a description of the network and address
// parameters.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	var addresses []string
	ip := net.ParseIP(host)
	if ip != nil {
		addresses = []string{address}
	} else {
		addresses, err = d.resolver().LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		for i, address := range addresses {
			addresses[i] = net.JoinHostPort(address, port)
		}
	}
	var primaries, fallbacks []string
	if d.dualStack() && network == "tcp" {
		primaries, fallbacks = partition(addresses)
	} else {
		primaries = addresses
	}

	var c net.Conn
	if len(fallbacks) > 0 {
		c, err = d.dialParallel(ctx, network, primaries, fallbacks)
	} else {
		c, err = d.dialSerial(ctx, network, primaries)
	}
	return c, err
}

func (d *Dialer) resolver() Resolver {
	if d.Resolver != nil {
		return d.Resolver
	}
	return net.DefaultResolver
}

func (d *Dialer) dualStack() bool {
	return d.FallbackDelay >= 0
}

func (d *Dialer) dialSerial(ctx context.Context, network string, addresses []string) (net.Conn, error) {
	var firstErr error
	for i, addr := range addresses {
		saddr := simpleAddr{addr: addr, network: network}
		select {
		case <-ctx.Done():
			return nil, &net.OpError{Op: "dial", Net: network, Source: d.LocalAddr, Addr: saddr, Err: ctx.Err()}
		default:
		}
		deadline, _ := ctx.Deadline()
		partialDeadline, err := partialDeadline(time.Now(), deadline, len(addresses)-i)
		if err != nil {
			// Ran out of time.
			if firstErr == nil {
				firstErr = &net.OpError{Op: "dial", Net: network, Source: d.LocalAddr, Addr: saddr, Err: err}
			}
			break
		}
		dialCtx := ctx
		if partialDeadline.Before(deadline) {
			var cancel context.CancelFunc
			dialCtx, cancel = context.WithDeadline(ctx, partialDeadline)
			defer cancel()
		}
		c, err := d.dialer().DialContext(dialCtx, network, addr)
		if err == nil {
			return c, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr == nil {
		firstErr = &net.OpError{Op: "dial", Net: network, Source: nil, Addr: nil, Err: errMissingAddress}
	}
	return nil, firstErr
}

// dialParallel races two copies of dialSerial, giving the first a
// head start. It returns the first established connection and
// closes the others. Otherwise it returns an error from the first
// primary address.
func (d *Dialer) dialParallel(ctx context.Context, network string, primaries, fallbacks []string) (net.Conn, error) {
	returned := make(chan struct{})
	defer close(returned)

	type dialResult struct {
		net.Conn
		error
		primary bool
		done    bool
	}
	results := make(chan dialResult) // unbuffered

	startRacer := func(ctx context.Context, primary bool) {
		ras := primaries
		if !primary {
			ras = fallbacks
		}
		c, err := d.dialSerial(ctx, network, ras)
		select {
		case results <- dialResult{Conn: c, error: err, primary: primary, done: true}:
		case <-returned:
			if c != nil {
				_ = c.Close()
			}
		}
	}

	var primary, fallback dialResult

	// Start the main racer.
	primaryCtx, primaryCancel := context.WithCancel(ctx)
	defer primaryCancel()
	go startRacer(primaryCtx, true)

	// Start the timer for the fallback racer.
	fallbackTimer := time.NewTimer(d.fallbackDelay())
	defer fallbackTimer.Stop()

	for {
		select {
		case <-fallbackTimer.C:
			fallbackCtx, fallbackCancel := context.WithCancel(ctx)
			defer fallbackCancel()
			go startRacer(fallbackCtx, false)

		case res := <-results:
			if res.error == nil {
				return res.Conn, nil
			}
			if res.primary {
				primary = res
			} else {
				fallback = res
			}
			if primary.done && fallback.done {
				return nil, primary.error
			}
			if res.primary && fallbackTimer.Stop() {
				// If we were able to stop the timer, that means it
				// was running (hadn't yet started the fallback), but
				// we just got an error on the primary path, so start
				// the fallback immediately (in 0 nanoseconds).
				fallbackTimer.Reset(0)
			}
		}
	}
}

func (d *Dialer) fallbackDelay() time.Duration {
	if d.FallbackDelay > 0 {
		return d.FallbackDelay
	} else {
		return 300 * time.Millisecond
	}
}

func (d *Dialer) dialer() *net.Dialer {
	if d.d != nil {
		return d.d
	}
	d.d = &net.Dialer{
		Timeout:       d.Timeout,
		Deadline:      d.Deadline,
		LocalAddr:     d.LocalAddr,
		FallbackDelay: d.FallbackDelay,
		KeepAlive:     d.KeepAlive,
		Control:       d.Control,
	}
	return d.d
}

// partialDeadline returns the deadline to use for a single address,
// when multiple addresses are pending.
func partialDeadline(now, deadline time.Time, addrsRemaining int) (time.Time, error) {
	if deadline.IsZero() {
		return deadline, nil
	}
	timeRemaining := deadline.Sub(now)
	if timeRemaining <= 0 {
		return time.Time{}, timeoutErr{}
	}
	// Tentatively allocate equal time to each remaining address.
	timeout := timeRemaining / time.Duration(addrsRemaining)
	// If the time per address is too short, steal from the end of the list.
	const saneMinimum = 2 * time.Second
	if timeout < saneMinimum {
		if timeRemaining < saneMinimum {
			timeout = timeRemaining
		} else {
			timeout = saneMinimum
		}
	}
	return now.Add(timeout), nil
}

// partition divides given address for dualstack usage
func partition(addresses []string) (primaries []string, fallbacks []string) {
	var primaryLabel bool
	for i, addr := range addresses {
		label := isIPv4(addr)
		if i == 0 || label == primaryLabel {
			primaryLabel = label
			primaries = append(primaries, addr)
		} else {
			fallbacks = append(fallbacks, addr)
		}
	}
	return
}

// isIPv4 reports whether addr contains an IPv4 address.
func isIPv4(addr string) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	return err == nil && tcpAddr.IP.To4() != nil
}
