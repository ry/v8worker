package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"

	"github.com/ry/v8worker"
)

type module struct {
	Err      error  `json:"err"`
	Source   string `json:"source"`
	Id       string `json:"id"`
	Filename string `json:"filename"`
	Dirname  string `json:"dirname"`
	main     bool
}

func (m *module) load() {
	filename := jsExtensionRe.ReplaceAllString(m.Id, "") + ".js"
	if wd, err := os.Getwd(); err == nil {
		m.Filename = path.Join(wd, filename)
	} else {
		m.Err = err
		return
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		m.Err = err
		return
	}
	m.Dirname = path.Dir(m.Filename)
	var b bytes.Buffer
	if m.main {
		b.WriteString(fmt.Sprintf(
			"var main = new NativeModule({ id: '%s', filename: '%s', dirname: '%s' });\n",
			m.Id, m.Filename, m.Dirname))
	}
	b.WriteString("(function (exports, require, module, __filename, __dirname) { ")
	if m.main {
		b.WriteString("\nrequire.main = module;")
	}
	b.Write(file)
	if m.main {
		b.WriteString("\n}")
		b.WriteString("(main.exports, NativeModule.require, main, main.filename, main.dirname));")
		b.WriteString("\n$send('exit');") // exit when main returns
	} else {
		b.WriteString("\n});")
	}
	m.Source = b.String()
}

// Adapted from node.js source:
// see https://github.com/nodejs/node/blob/master/src/node.js#L871
const nativeModule = `
	'use strict';

	function NativeModule(rawModule) {
		this.filename = rawModule.filename;
		this.dirname = rawModule.dirname;
		this.id = rawModule.id;
		this.exports = {};
		this.loaded = false;
		this._source = rawModule.source;
	}

	NativeModule.require = function(id) {
		var rawModule = JSON.parse($sendSync(id));
		if (rawModule.err) {
			throw new RangeError(JSON.stringify(rawModule.err));
		}

		var nativeModule = new NativeModule(rawModule);

		nativeModule.compile();

		return nativeModule.exports;
	};

	NativeModule.prototype.compile = function() {
		var fn = eval(this._source);
		fn(this.exports, NativeModule.require, this, this.filename, this.dirname);
		this.loaded = true;
	};
	`

var (
	jsExtensionRe = regexp.MustCompile(`\.js$`)
	jsFile        = flag.String("f", "module-1.js", "js file to run")
)

func loadMainModule(w *v8worker.Worker, id string) error {
	m := module{Id: id, main: true}
	m.load()
	if m.Err != nil {
		return m.Err
	}
	return w.Load(m.Filename, m.Source)
}

func runWorker(done <-chan struct{}, scriptFile string) <-chan string {
	out := make(chan string)

	go func() {
		worker := v8worker.New(func(msg string) {
			out <- msg
		}, func(msg string) string {
			m := module{Id: msg, main: false}
			m.load()
			bytes, _ := json.Marshal(m)
			return string(bytes)
		})

		defer func() {
			close(out)
			worker.TerminateExecution()
		}()

		if err := worker.Load("native-module.js", nativeModule); err != nil {
			log.Println(err)
			return
		}
		if err := loadMainModule(worker, scriptFile); err != nil {
			log.Println(err)
			return
		}

		for {
			select {
			case <-done:
				log.Println("killing worker...")
				return
			default:
			}
		}
	}()

	return out
}

func main() {
	flag.Parse()
	done := make(chan struct{})
	in := runWorker(done, *jsFile)

	for msg := range in {
		if msg == "exit" {
			log.Println("got exit message from js...")
			close(done)
		} else {
			log.Printf("message from js: %q\n", msg)
		}
	}
}
