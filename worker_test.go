package v8worker

import (
	"testing"
)

func TestVersion(t *testing.T) {
	println(Version())
}

func TestBasic(t *testing.T) {
	recvCount := 0
	worker := New(func(msg string) {
		println("recv cb", msg)
		if msg != "hello" {
			t.Fatal("bad msg", msg)
		}
		recvCount++
	})

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

func TestMultipleWorkers(t *testing.T) {
	recvCount := 0
	worker1 := New(func(msg string) {
		println("w1", msg)
		recvCount++
	})
	worker2 := New(func(msg string) {
		println("w2", msg)
		recvCount++
	})

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
