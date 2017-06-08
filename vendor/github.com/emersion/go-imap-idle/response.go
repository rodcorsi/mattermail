package idle

import (
	"github.com/emersion/go-imap"
)

// An IDLE response.
type Response struct {
	Done   <-chan struct{}
	Writer *imap.Writer
}

func (r *Response) HandleFrom(hdlr imap.RespHandler) error {
	// Wait for a continuation request
	for h := range hdlr {
		if _, ok := h.Resp.(*imap.ContinuationResp); ok {
			h.Accept()
			break
		}
		h.Reject()
	}

	// We got a continuation request, ignore all responses and wait for r.Done to
	// be closed
	for {
		select {
		case h, more := <-hdlr:
			if !more {
				return nil
			}
			h.Reject()
		case <-r.Done:
			if _, err := r.Writer.Write([]byte(doneLine + "\r\n")); err != nil {
				return err
			}
			if err := r.Writer.Flush(); err != nil {
				return err
			}
			return nil
		}
	}
}
