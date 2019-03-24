package nexus

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ScriptTypeGroovy is used when specifying Type of groovy scripts
var ScriptTypeGroovy = String("groovy")

// Script represents a Nexus groovy script.
// This interface can be used for executing ad-hoc groovy logic on the host.
// Some of the exported Nexus methods use this interface on the backend to compensate
// for non-existant REST functionality. When the script raises an error, err is
// populated with the text of the exception raised in the groovy environment.
//
// Args can be provided to the execution in any structure that can be marshalled to JSON.
// If the script takes no arguments, then specify nil. To consume these arguments within the script,
// it could start with something like this:
//
//     import groovy.json.JsonSlurper
//     parsed_args = new JsonSlurper().parseText(args)
//     return parsed_args.name // from an input like {"name": "Bob"}
//
// Example 1
//
// Execute a "hello world" script, log the result, then remove the script
//    func main() {
//      client, err := nexus.New("http://localhost:8081", "admin", "admin123")
//      if err != nil {
//        log.Fatal(err)
//      }
//      script, err := client.CreateScript(&nexus.Script{
//        Name:    nexus.String("my-script"),
//        Type:    nexus.ScriptTypeGroovy,
//        Content: nexus.String("return 'Hello World'"),
//      })
//      if err != nil {
//        log.Fatal(err)
//      }
//      res, err := script.Execute(nil)
//      if err != nil {
//        log.Fatal(err)
//      }
//      log.Println("Result: ", *res.Result)
//      err = script.Delete()
//      log.Println(err)
//    }
//    // 2019/03/23 10:39:07 Result:  Hello World
//    // 2019/03/23 10:39:07 <nil>
//
// Example 2
//
// Execute a script that would raise an exception
//      script, err := client.CreateScript(&nexus.Script{
//        Name:    nexus.String("my-script"),
//        Type:    nexus.ScriptTypeGroovy,
//        Content: nexus.String("gah"),
//      })
//      if err != nil {
//        log.Fatal(err)
//      }
//      _, err = script.Execute(nil)
//      log.Println("Error: ", err.Error())
//      _ = script.Delete()
//
//    // 2019/03/23 15:24:42 Error:  javax.script.ScriptException: groovy.lang.MissingPropertyException: No such property: gah for class: Script52
type Script struct {
	Name    *string `json:"name"`
	Content *string `json:"content"`
	Type    *string `json:"type"`

	client *Nexus
}

// EphemeralScript is the same as Script, but provides an Execute() call which
// will create and destroy a throw-away script for you after execution.
type EphemeralScript struct {
	*Script
}

// NewEphemeralScript takes a script instance where only the Content is required,
// and returns an executable instance.
func (n *Nexus) NewEphemeralScript(script *Script) (boundScript *EphemeralScript) {
	boundScript = &EphemeralScript{
		Script: script,
	}
	boundScript.Script.client = n
	return
}

// Execute creates, executes, and then destroys an ephemeral script.
func (s *EphemeralScript) Execute(args interface{}) (res *ExecuteScriptResponse, err error) {
	script, err := s.Script.client.CreateScript(&Script{
		Name:    String(uuid.New().String()),
		Type:    ScriptTypeGroovy,
		Content: s.Script.Content,
	})
	if err != nil {
		return
	}
	defer script.Delete()
	res, err = script.Execute(args)
	return
}

// ListScriptsResponse contains a collection of scripts from a ListScripts call
type ListScriptsResponse struct {
	Scripts []*Script
}

// ExecuteScriptResponse contains the contents of the response from executing a script
type ExecuteScriptResponse struct {
	Name   *string `json:"name"`
	Result *string `json:"result"`
}

// Execute this script instance
func (s *Script) Execute(args interface{}) (res *ExecuteScriptResponse, err error) {
	res, err = s.client.ExecuteScript(*s.Name, args)
	return
}

// Delete this script instance
func (s *Script) Delete() (err error) {
	err = s.client.DeleteScript(*s.Name)
	return
}

// ensureAndExecute is used internally for the process of ensuring the contents
// of a script and subsequently executing it.
func (s *Script) ensureAndExecute(args interface{}) (res *ExecuteScriptResponse, err error) {
	script, err := s.client.GetScript(*s.Name)
	if err != nil {
		script, err = s.client.CreateScript(&Script{
			Name:    s.Name,
			Type:    s.Type,
			Content: s.Content,
		})
		if err != nil {
			return
		}
	}
	if *script.Content != *s.Content {
		script, err = s.client.UpdateScript(&Script{
			Name:    s.Name,
			Type:    s.Type,
			Content: s.Content,
		})
		if err != nil {
			return
		}
	}
	res, err = script.Execute(args)
	return
}

// GetScript returns an executable script instance by name
func (n *Nexus) GetScript(name string) (script *Script, err error) {
	url := fmt.Sprintf("service/rest/v1/script/%s", name)
	req, err := n.NewRequest("GET", url, nil, nil, "")
	if err != nil {
		return
	}
	body, err := n.Do(req, map[int]string{
		404: fmt.Sprintf("Script %s does not exist", name),
	}, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &script)
	if err != nil {
		return
	}
	script.client = n
	return
}

func marshalScriptArgs(args interface{}) (payload []byte, err error) {
	if args == nil {
		payload = nil
	} else {
		payload, err = json.Marshal(args)
	}
	return
}

// ExecuteScript executes the script with the given name and returns the result.
// Args must be a structure that can be marshaled to JSON or nil.
func (n *Nexus) ExecuteScript(name string, args interface{}) (res *ExecuteScriptResponse, err error) {
	payload, err := marshalScriptArgs(args)
	if err != nil {
		return
	}
	url := fmt.Sprintf("service/rest/v1/script/%s/run", name)
	req, err := n.NewRequest("POST", url, nil, payload, "text/plain")
	if err != nil {
		return
	}
	body, err := n.Do(req, nil, true)
	if err != nil {
		err = json.Unmarshal([]byte(err.Error()), &res)
		if err != nil {
			return
		}
		err = errors.New(*res.Result)
		return
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return
	}
	return
}

// DeleteScript deletes a script with a given name
func (n *Nexus) DeleteScript(name string) (err error) {
	url := fmt.Sprintf("service/rest/v1/script/%s", name)
	req, err := n.NewRequest("DELETE", url, nil, nil, "")
	if err != nil {
		return
	}
	_, err = n.Do(req, map[int]string{
		404: fmt.Sprintf("Script %s does not exist", name),
	}, false)
	return
}

// ListScripts returns all script objects on the Nexus host
func (n *Nexus) ListScripts() (res *ListScriptsResponse, err error) {
	res = &ListScriptsResponse{}
	req, err := n.NewRequest("GET", "service/rest/v1/script", nil, nil, "")
	if err != nil {
		return
	}
	body, err := n.Do(req, nil, false)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &res.Scripts)
	if err != nil {
		return
	}
	for _, x := range res.Scripts {
		x.client = n
	}
	return
}

// CreateScript creates a new script with the given parameters and returns
// a copy of the provided instance with the bound client so Execute() can be called on it.
func (n *Nexus) CreateScript(script *Script) (boundScript *Script, err error) {
	if script.Name == nil || script.Content == nil {
		err = errors.New("Script instance must contain a name and content")
		return
	}
	payload, err := json.Marshal(script)
	if err != nil {
		return
	}
	req, err := n.NewRequest("POST", "service/rest/v1/script", nil, payload, "")
	if err != nil {
		return
	}
	_, err = n.Do(req, map[int]string{
		500: fmt.Sprintf("Script with name %s already exists, use UpdateScript instead", *script.Name),
	}, false)
	if err != nil {
		return
	}
	boundScript = &Script{
		Name:    script.Name,
		Type:    script.Type,
		Content: script.Content,
		client:  n,
	}
	return
}

// UpdateScript takes the given script instance and ensures it's counterpart on Nexus
// by the same name has the same contents.
func (n *Nexus) UpdateScript(script *Script) (boundScript *Script, err error) {
	if script.Name == nil || script.Content == nil {
		err = errors.New("Script instance must contain a name and content")
		return
	}
	payload, err := json.Marshal(script)
	if err != nil {
		return
	}
	url := fmt.Sprintf("service/rest/v1/script/%s", *script.Name)
	req, err := n.NewRequest("PUT", url, nil, payload, "")
	if err != nil {
		return
	}
	_, err = n.Do(req, map[int]string{
		404: fmt.Sprintf("Script with name %s doesn't exist", *script.Name),
	}, false)
	if err != nil {
		return
	}
	boundScript = &Script{
		Name:    script.Name,
		Type:    script.Type,
		Content: script.Content,
		client:  n,
	}
	return
}
