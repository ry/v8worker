package v8worker

import (
	"runtime"
	"testing"
	"time"
)

func TestVersion(t *testing.T) {
	println(Version())
}

func DiscardSendSync(msg string) string { return "" }

func TestBasic(t *testing.T) {
	recvCount := 0
	worker := New(func(msg string) {
		println("recv cb", msg)
		if msg != "hello" {
			t.Fatal("bad msg", msg)
		}
		recvCount++
	}, DiscardSendSync)

	code := ` $print("ready"); `
	err := worker.Load("code.js", code)
	if err != nil {
		t.Fatal(err)
	}

	codeWithSyntaxError := ` $print(hello world"); `
	err = worker.Load("codeWithSyntaxError.js", codeWithSyntaxError)
	if err == nil {
		t.Fatal("Expected error")
	}
	//println(err.Error())

	codeWithRecv := `
		$recv(function(msg) {
			$print("recv msg", msg);
		});
		$print("ready");
	`
	err = worker.Load("codeWithRecv.js", codeWithRecv)
	if err != nil {
		t.Fatal(err)
	}
	worker.Send("hi")

	codeWithSend := `
		$send("hello");
		$send("hello");
	`
	err = worker.Load("codeWithSend.js", codeWithSend)
	if err != nil {
		t.Fatal(err)
	}

	if recvCount != 2 {
		t.Fatal("bad recvCount", recvCount)
	}
}

func TestUint8Array(t *testing.T) {
	worker := New(func(msg string) {}, DiscardSendSync)
	codeWithArrayBufferAllocator := ` var uint8 = new Uint8Array(256); $print(uint8); `
	err := worker.Load("buffer.js", codeWithArrayBufferAllocator)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMultipleWorkers(t *testing.T) {
	recvCount := 0
	worker1 := New(func(msg string) {
		println("w1", msg)
		recvCount++
	}, DiscardSendSync)
	worker2 := New(func(msg string) {
		println("w2", msg)
		recvCount++
	}, DiscardSendSync)

	err := worker1.Load("1.js", `$send("hello1")`)
	if err != nil {
		t.Fatal(err)
	}

	err = worker2.Load("2.js", `$send("hello2")`)
	if err != nil {
		t.Fatal(err)
	}

	if recvCount != 2 {
		t.Fatal("bad recvCount", recvCount)
	}
}

func TestRequestFromJS(t *testing.T) {
	var caught string
	worker := New(func(msg string) {
		println("recv cb", msg)
		caught = msg
	}, func(msg string) string {
		println("send sync exchange", msg)
		return msg + " exchanged"
	})
	code := `
	var response = $sendSync("ping");
	$send(response);
`
	err := worker.Load("code.js", code)
	if err != nil {
		t.Fatal(err)
	}
	if caught != "ping exchanged" {
		t.Fail()
	}
}

func TestRequestFromGo(t *testing.T) {
	var caught string
	worker := New(func(msg string) {
		println("recv cb", msg)
		caught = msg
	}, DiscardSendSync)
	code := `
	$recvSync(function(msg) {
		$send("in recvSync:"+msg);
		return msg + " exchanged";
	});
`
	err := worker.Load("code.js", code)
	if err != nil {
		t.Fatal(err)
	}
	response := worker.SendSync("pong")
	if got, want := response, "pong exchanged"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRequestFromGoReturningNonString(t *testing.T) {
	worker := New(func(msg string) {
		println("recv cb", msg)
	}, DiscardSendSync)
	code := `
	$recvSync(function(msg) {
		$send("in recvSync:"+msg);
		return 42;
	});
`
	err := worker.Load("code.js", code)
	if err != nil {
		t.Fatal(err)
	}
	response := worker.SendSync("pang")
	if got, want := response, "err: non-string return value"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

//I have profiled this repeatedly with massive values to ensure
//memory does indeed get reclaimed and that the finalizer
// gets called and the C-side of this does clean up memory correctly.
func TestWorkerDeletion(t *testing.T) {
	recvCount := 0
	for i := 1; i <= 100; i++ {
		worker := New(func(msg string) {
			println("worker", msg)
			recvCount++
		}, DiscardSendSync)
		err := worker.Load("1.js", `$send("hello1")`)
		if err != nil {
			t.Fatal(err)
		}
		runtime.GC()
	}

	if recvCount != 100 {
		t.Fatal("bad recvCount", recvCount)
	}
}

// Test breaking script execution
func TestWorkerBreaking(t *testing.T) {
	worker := New(func(msg string) {
		println("recv cb", msg)
	}, DiscardSendSync)

	go func(w *Worker) {
		time.Sleep(time.Second)
		w.TerminateExecution()
	}(worker)

	worker.Load("forever.js", ` while (true) { ; } `)
}

func TestTightCreateLoop(t *testing.T) {
	println("create 3000 workers in tight loop to see if we get OOM")
	for i := 0; i < 3000; i++ {
		runSimpleWorker(t)
	}
	println("success")
}

func runSimpleWorker(t *testing.T) {
	w := New(nil, nil)
	defer w.Dispose()

	err := w.Load("mytest.js", `
	               // Do something
	               var something = "Simple JavaScript";
	       `)

	if err != nil {
		t.Fatal(err)
	}
}
