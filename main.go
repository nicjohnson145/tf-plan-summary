package main

import (
	"flag"
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"sort"
	"strings"
)

const (
	No_Op = "no-op"
	Create = "create"
	Read = "read"
	Update = "update"
	Delete = "delete"

	// Pseudo state
	Recreate = "recreate"
)

var (
	excludeReads bool
	noTrimPrefix bool

	no_op = []string{No_Op}
	create = []string{Create}
	read = []string{Read}
	update = []string{Update}
	deleteCreate = []string{Delete, Create}
	createDelete = []string{Create, Delete}
	delete_ = []string{Delete}
)

func init() {
	usage := "Exclude read operations from being displayed"
	flag.BoolVar(&excludeReads, "exclude-reads", false, usage)
	flag.BoolVar(&excludeReads, "x", false, usage+" (shorthand)")
	flag.BoolVar(&noTrimPrefix, "no-trim", false, "don't trim the common prefix from addresses")
}

type Change struct {
	Actions []string `json:"actions"`
}

type ChangeRepr struct {
	Address string `json:"address"`
	Change  Change `json:"change"`
}

type Plan struct {
	ResourceChanges []ChangeRepr `json:"resource_changes"`
}

type Output struct {
	Address string
	Type string
}

func (o Output) String() string {
	return fmt.Sprintf("(%8v) %v", o.Type, o.Address)
}

func ok(err error, msg string) {
	if err != nil {
		fmt.Printf("%v: %v", msg, err)
		os.Exit(1)
	}
}

func arrayEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func getOutput(change ChangeRepr) *Output {
	o := &Output{Address: change.Address}
	switch {
	case arrayEqual(change.Change.Actions, no_op):
		o.Type = No_Op
	case arrayEqual(change.Change.Actions, create):
		o.Type = Create
	case arrayEqual(change.Change.Actions, read):
		o.Type = Read
	case arrayEqual(change.Change.Actions, update):
		o.Type = Update
	case arrayEqual(change.Change.Actions, deleteCreate):
		o.Type = Recreate
	case arrayEqual(change.Change.Actions, createDelete):
		o.Type = Recreate
	case arrayEqual(change.Change.Actions, delete_):
		o.Type = Delete
	default:
		fmt.Printf("Unknown change sequence of %v for %v", change.Change.Actions, change.Address)
		return nil
	}
	return o
}

func trimPrefix(changes []*Output) {
	addresses := []string{}
	for _, c := range changes {
		addresses = append(addresses, c.Address)
	}

	sort.Strings(addresses)

	if len(addresses) < 2 {
		return
	}

	charCount := len(addresses[0])
	if len(addresses[1]) < len(addresses[0]) {
		charCount = len(addresses[1])
	}

	prefix := ""
	for i := 0; i < charCount; i++ {
		if addresses[0][i] == addresses[1][i] {
			prefix = prefix + fmt.Sprintf("%c", addresses[0][i])
		} else {
			break
		}
	}

	for _, c := range changes {
		c.Address = strings.TrimPrefix(c.Address, prefix)
	}
}

func main() {
	flag.Parse()

	bytes, err := ioutil.ReadAll(os.Stdin)
	ok(err, "Error reading from stdin")

	var plan Plan
	err = json.Unmarshal(bytes, &plan)
	ok(err, "Error unmarshalling")

	changes := []*Output{}
	for _, change := range plan.ResourceChanges {
		output := getOutput(change)

		if output == nil {
			continue
		}

		if output.Type == No_Op {
			continue
		}

		if output.Type == Read && excludeReads {
			continue
		}

		changes = append(changes, output)
	}

	if !noTrimPrefix {
		trimPrefix(changes)
	}

	for _, change := range changes {
		fmt.Println(change)
	}
}
