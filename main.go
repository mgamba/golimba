package main

import (
	"fmt"
	"io"
  "io/ioutil"
	"os"
  "regexp"
  "strconv"

	"github.com/hajimehoshi/oto"
	"github.com/hajimehoshi/go-mp3"

	prompt "github.com/c-bata/go-prompt"
)

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

var mode = ""
var liveArg = ""
var loadMatch = regexp.MustCompile(`^load`)

func executor(in string) {
	if LivePrefixState.IsEnable {
    fmt.Println("executing on " + in)
    switch mode {
    case "load":
      fname := liveArg
      slot := in
      fmt.Printf("Loading %v into slot %v\n", fname, slot)
    }
    mode = ""
		LivePrefixState.IsEnable = false
		LivePrefixState.LivePrefix = ""
    return
  }
	if in == "" {
		LivePrefixState.IsEnable = false
		LivePrefixState.LivePrefix = in
		return
	}
	LivePrefixState.LivePrefix = in + "> "
	LivePrefixState.IsEnable = true
  switch {
  case loadMatch.MatchString(in):
    mode = "load"
    liveArg = in[5:len(in)-1]
  }
}

func completer(in prompt.Document) []prompt.Suggest {
  s := []prompt.Suggest{}
  if len(in.Text) > 0 && !LivePrefixState.IsEnable {
    if _, err := strconv.Atoi(in.Text); err == nil {
      go run()
    } else {
      switch {
      case loadMatch.MatchString(in.Text):
        files, err := ioutil.ReadDir("./")
        if err != nil {
          panic(err)
        }
        for _, f := range files {
          s = append(s, prompt.Suggest{Text: f.Name(), Description: ""})
        }
      default:
        s = []prompt.Suggest{
          {Text: "slice", Description: "enter slice mode"},
          {Text: "play", Description: "enter play mode"},
          {Text: "load", Description: "load file"},
        }
      }
    }
  }
  return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}

func main() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionTitle("live-prefix-example"),
	)
	p.Run()
}










var (
	c, _ = oto.NewContext(44100, 2, 2, 8192)
)

func run() error {
	f, err := os.Open("kit.mp3")
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}

	p := c.NewPlayer()
	if err != nil {
		return err
	}
	defer p.Close()

	if _, err := io.CopyN(p, d, 44000); err != nil {
		return err
	}
	return nil
}
