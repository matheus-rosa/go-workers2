package workers

import (
	"github.com/customerio/gospec"
	. "github.com/customerio/gospec"
)

var middlewareCalled bool

type testMiddleware struct{}

func (l *testMiddleware) Call(queue string, message interface{}, next func()) {
	middlewareCalled = true
	next()
}

func WorkerSpec(c gospec.Context) {
	var processed = make([]string, 0)
	middlewareCalled = false

	var testJob = (func(message interface{}) bool {
		processed = append(processed, message.(string))
		return true
	})

	manager := newManager("myqueue", testJob, 1)

	c.Specify("newWorker", func() {
		c.Specify("it returns an instance of worker with connection to manager", func() {
			worker := newWorker(manager)
			c.Expect(worker.manager, Equals, manager)
		})
	})

	c.Specify("work", func() {
		worker := newWorker(manager)

		c.Specify("gives each message to a new job instance, and calls perform", func() {
			messages := make(chan string)
			go worker.work(messages)
			messages <- "test"

			c.Expect(len(processed), Equals, 1)
			c.Expect(processed, Contains, "test")

			close(messages)
		})

		c.Specify("confirms job completed", func() {
			messages := make(chan string)
			go worker.work(messages)
			messages <- "test"

			c.Expect(<-manager.confirm, Equals, "test")

			close(messages)
		})

		c.Specify("runs defined middleware", func() {
			Middleware.Append(&testMiddleware{})

			messages := make(chan string)
			go worker.work(messages)
			messages <- "test"

			c.Expect(middlewareCalled, IsTrue)

			close(messages)
		})
	})
}
