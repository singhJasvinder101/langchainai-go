package llm

// Stream turns the channel pair returned by Provider.GenerateStream into a
// pull-based iterator, in the style of bufio.Scanner or sql.Rows: call Next
// in a loop, read the current chunk via Current (or Text as a shortcut), and
// check Err once the loop ends.
//
//	stream := llm.NewStream(provider.GenerateStream(ctx, req))
//	for stream.Next() {
//		fmt.Print(stream.Text())
//	}
//	if err := stream.Err(); err != nil {
//		log.Fatal(err)
//	}
type Stream struct {
	responses <-chan *StreamResponse
	errs      <-chan error
	current   *StreamResponse
	err       error
}

// NewStream wraps the (responses, errs) channel pair returned by
// Provider.GenerateStream.
func NewStream(responses <-chan *StreamResponse, errs <-chan error) *Stream {
	return &Stream{responses: responses, errs: errs}
}

// Next advances the stream to the next chunk and reports whether one was
// read. It returns false once the stream is exhausted or an error occurred —
// call Err after the loop to tell the two apart.
func (s *Stream) Next() bool {
	for s.responses != nil || s.errs != nil {
		select {
		case chunk, ok := <-s.responses:
			if !ok {
				s.responses = nil
				continue
			}
			s.current = chunk
			return true
		case err, ok := <-s.errs:
			if !ok {
				s.errs = nil
				continue
			}
			if err != nil {
				s.err = err
				return false
			}
		}
	}
	return false
}

// Current returns the chunk read by the most recent call to Next.
func (s *Stream) Current() *StreamResponse {
	return s.current
}

// Text is a shortcut for Current().Text().
func (s *Stream) Text() string {
	return s.current.Text()
}

// Err returns the first error encountered, once Next has returned false
// because the stream failed rather than completed normally.
func (s *Stream) Err() error {
	return s.err
}
