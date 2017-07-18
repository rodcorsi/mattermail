package idle

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/responses"
)

// An IDLE response.
type Response struct {
	Stop   <-chan struct{}
	Done   chan<- error
	Writer *imap.Writer

	gotContinuationReq bool
}

func (r *Response) stop() error {
	if _, err := r.Writer.Write([]byte(doneLine + "\r\n")); err != nil {
		return err
	}
	return r.Writer.Flush()
}

func (r *Response) Handle(resp imap.Resp) error {
	// Wait for a continuation request
	if _, ok := resp.(*imap.ContinuationReq); ok && !r.gotContinuationReq {
		r.gotContinuationReq = true

		// We got a continuation request, wait for r.Stop to be closed
		go func() {
			<-r.Stop
			r.Done <- r.stop()
		}()

		return nil
	}

	return responses.ErrUnhandled
}
