*Considerations for the future*

From context documentation, this type of context handling may be made
available in the wrapper in the future, but with go1.7+ including
context the methods, interface may be cleaner and ubiquitous

func(){
  // Pass a context with a timeout to tell a blocking function that it
  // should abandon its work after the timeout elapses.
  ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

  select {
  case <-time.After(1 * time.Second):
      fmt.Println("overslept")
  case <-ctx.Done():
      fmt.Println(ctx.Err()) // prints "context deadline exceeded"
  }

  // Even though ctx should have expired already, it is good
  // practice to call its cancelation function in any case.
  // Failure to do so may keep the context and its parent alive
  // longer than necessary.
  cancel()
}

type WithContext func(context.Context, http.Handler) http.Handler

// ServeHTTP satisies the http.Handler interface.
func (fn WithContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fn(ctx, w, r)
}
