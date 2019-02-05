package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"

	"github.com/textileio/textile-go/core"
	"github.com/textileio/textile-go/util"
)

var errMissingPeerId = errors.New("missing peer id")
var errMissingSearchInfo = errors.New("missing search info")

func init() {
	register(&contactsCmd{})
}

type contactsCmd struct {
	Add    addContactsCmd  `command:"add" description:"Add a contact"`
	Ls     lsContactsCmd   `command:"ls" description:"List known contacts"`
	Get    getContactsCmd  `command:"get" description:"Get a known contact"`
	Remove rmContactsCmd   `command:"rm" description:"Remove a known contact"`
	Find   findContactsCmd `command:"find" description:"Find contacts"`
}

func (x *contactsCmd) Name() string {
	return "contacts"
}

func (x *contactsCmd) Short() string {
	return "Manage contacts"
}

func (x *contactsCmd) Long() string {
	return `
Use this command to add, list, get, and remove local contacts and find other contacts on the network.
`
}

// TODO: make this work
type addContactsCmd struct {
	Client ClientOptions `group:"Client Options"`
}

func (x *addContactsCmd) Usage() string {
	return `

Add to known contacts.
`
}

func (x *addContactsCmd) Execute(args []string) error {
	setApi(x.Client)
	res, err := executeStringCmd(PUT, "contacts", params{})
	if err != nil {
		return err
	}
	output(res)
	return nil
}

type lsContactsCmd struct {
	Client ClientOptions `group:"Client Options"`
	Thread string        `short:"t" long:"thread" description:"Thread ID. Omit for all known contacts."`
}

func (x *lsContactsCmd) Usage() string {
	return `

Lists known contacts.
Include the --thread flag to list contacts for a given thread.`
}

func (x *lsContactsCmd) Execute(args []string) error {
	setApi(x.Client)
	var list []core.ContactInfo
	res, err := executeJsonCmd(GET, "contacts", params{
		opts: map[string]string{
			"thread": x.Thread,
		},
	}, &list)
	if err != nil {
		return err
	}
	output(res)
	return nil
}

type getContactsCmd struct {
	Client ClientOptions `group:"Client Options"`
}

func (x *getContactsCmd) Usage() string {
	return `

Gets a known contact.`
}

func (x *getContactsCmd) Execute(args []string) error {
	setApi(x.Client)
	if len(args) == 0 {
		return errMissingPeerId
	}
	var info core.ContactInfo
	res, err := executeJsonCmd(GET, "contacts/"+args[0], params{}, &info)
	if err != nil {
		return err
	}
	output(res)
	return nil
}

type rmContactsCmd struct {
	Client ClientOptions `group:"Client Options"`
}

func (x *rmContactsCmd) Usage() string {
	return `

Removes a known contact.`
}

func (x *rmContactsCmd) Execute(args []string) error {
	setApi(x.Client)
	if len(args) == 0 {
		return errMissingPeerId
	}
	res, err := executeStringCmd(DEL, "contacts/"+args[0], params{})
	if err != nil {
		return err
	}
	output(res)
	return nil
}

type findContactsCmd struct {
	Client   ClientOptions `group:"Client Options"`
	Username string        `short:"u" long:"username" description:"A username to use in the search."`
	Peer     string        `short:"p" long:"peer" description:"A Peer ID use in the search."`
	Address  string        `short:"a" long:"address" description:"An account address to use in the search."`
	Local    bool          `long:"local" description:"Only search local contacts."`
	Limit    int           `long:"limit" description:"Stops searching after limit results are found." default:"5"`
	Wait     int           `long:"wait" description:"Stops searching after 'wait' seconds have elapsed." default:"5"`
}

func (x *findContactsCmd) Usage() string {
	return `

Finds contacts known locally and on the network.
`
}

func (x *findContactsCmd) Execute(args []string) error {
	setApi(x.Client)
	if x.Username == "" && x.Peer == "" && x.Address == "" {
		return errMissingSearchInfo
	}

	resultsCh := make(chan core.ContactQueryResult)
	outputCh := make(chan interface{})
	doneCh := make(chan struct{})

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		defer close(resultsCh)

		res, cancel, err := request(POST, "contacts/search", params{
			opts: map[string]string{
				"username": x.Username,
				"peer":     x.Peer,
				"address":  x.Address,
				"local":    strconv.FormatBool(x.Local),
				"limit":    strconv.Itoa(x.Limit),
				"wait":     strconv.Itoa(x.Wait),
			},
		})
		if err != nil {
			outputCh <- err.Error()
			return
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 {
			body, err := util.UnmarshalString(res.Body)
			if err != nil {
				outputCh <- err.Error()
			} else {
				outputCh <- body
			}
			return
		}

		decoder := json.NewDecoder(res.Body)
		for {
			select {
			case <-doneCh:
				outputCh <- ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
				cancel()
				return
			default:
				var result core.ContactQueryResult
				if err := decoder.Decode(&result); err == io.EOF {
					outputCh <- ">>>sdfvadsfo;v;odfnvo;dfnv>"
					quit <- os.Interrupt
					break
				} else if err != nil {
					outputCh <- err.Error()
					return
				}
				resultsCh <- result
			}
		}
	}()

	go func() {
		for {
			select {
			case res, ok := <-resultsCh:
				if !ok {
					return
				}

				data, err := json.MarshalIndent(res, "", "    ")
				if err == io.EOF {
					break
				} else if err != nil {
					return
				}
				outputCh <- string(data)
			}
		}
	}()

	for {
		select {
		case val := <-outputCh:
			output(val)

		case <-quit:
			fmt.Println("Interrupted")
			fmt.Printf("Canceling...")
			doneCh <- struct{}{}
			fmt.Print("done\n")
			os.Exit(1)
			return nil
		}
	}
}
