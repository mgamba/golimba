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
var trimMatch = regexp.MustCompile(`^trim`)
var seekMatch = regexp.MustCompile(`^seek`)
var fileSlots = [10]string{}
var sampleStarts = [10]int{}
var sampleLengths = [10]int64{}

func executor(in string) {
	if LivePrefixState.IsEnable {
    fmt.Println("executing on " + in)
    switch mode {
    case "load":
      fname := liveArg
      slot, _ := strconv.Atoi(in)
      fmt.Printf("Loading %v into slot %v\n", fname, slot)
      fileSlots[slot] = fname
      sampleStarts[slot] = 0
      sampleLengths[slot] = 44100
    case "trim":
      if len(in) > 0 {
        slot, _ := strconv.Atoi(liveArg)
        amount, _ := strconv.Atoi(in)
        fmt.Printf("Changing slot %v by amount %v\n", slot, amount)
        sampleLengths[slot] = max(sampleLengths[slot] + int64(amount), 0)
        run(slot)
        return
      }
    case "seek":
      if len(in) > 0 {
        slot, _ := strconv.Atoi(liveArg)
        amount, _ := strconv.Atoi(in)
        fmt.Printf("Changing slot %v start by amount %v\n", slot, amount)
        sampleStarts[slot] += amount
        run(slot)
        return
      }
    }
    mode = ""
		LivePrefixState.IsEnable = false
		LivePrefixState.LivePrefix = ""
    return
  }
  switch {
  case loadMatch.MatchString(in):
    mode = "load"
    liveArg = in[5:len(in)]
  case trimMatch.MatchString(in):
    mode = "trim"
    liveArg = in[5:len(in)]
  case seekMatch.MatchString(in):
    mode = "seek"
    liveArg = in[5:len(in)]
  default:
		LivePrefixState.IsEnable = false
		LivePrefixState.LivePrefix = ""
		return
  }
	LivePrefixState.LivePrefix = in + "> "
	LivePrefixState.IsEnable = true
}

func completer(in prompt.Document) []prompt.Suggest {
  s := []prompt.Suggest{}
  if len(in.Text) > 0 && !LivePrefixState.IsEnable {
    if slot, err := strconv.Atoi(string(in.Text[len(in.Text)-1])); err == nil {
      go run(slot)
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
      case trimMatch.MatchString(in.Text) || seekMatch.MatchString(in.Text):
        for i := 0; i <= 9; i++ {
          s = append(s, prompt.Suggest{Text: strconv.Itoa(i), Description: ""})
        }
      default:
        s = []prompt.Suggest{
          {Text: "slice", Description: "enter slice mode"},
          {Text: "play", Description: "enter play mode"},
          {Text: "load", Description: "load file"},
          {Text: "trim", Description: "trim length"},
          {Text: "seek", Description: "sample start"},
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

func run(slot int) error {
  fname := fileSlots[slot]
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}

  d.Seek(min(int64(sampleStarts[slot]), d.Length()-1), 0)

	p := c.NewPlayer()
	if err != nil {
		return err
	}
	defer p.Close()

	if _, err := io.CopyN(p, d, sampleLengths[slot]); err != nil {
		return err
	}
	return nil
}

func min(x, y int64) int64 {
    if x > y {
        return y
    }
    return x
}

func max(x, y int64) int64 {
    if x < y {
        return y
    }
    return x
}
