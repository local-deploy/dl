package teleport

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/local-deploy/dl/project"
	"github.com/m7shapan/njson"
)

type teleport struct {
	Proxy   string
	User    string
	Node    string
	Catalog string
}

type status struct {
	Cluster string   `njson:"active.cluster"`
	Logins  []string `njson:"active.logins"`
}

type node struct {
	Node string `njson:"spec.hostname"`
}

type nodes []node

var s status
var nn nodes

func getClient() (t *teleport, err error) {
	err = setTeleportStatus()
	if err != nil {
		return nil, err
	}

	err = setTeleportNodes()
	if err != nil {
		return nil, err
	}

	env := project.Env.GetString("TELEPORT")
	u := strings.Split(env, ":")

	c := &teleport{
		Proxy:   s.Cluster,
		User:    u[0],
		Node:    u[1],
		Catalog: project.Env.GetString("CATALOG_SRV"),
	}

	err = checkAccess(c)
	if err != nil {
		return nil, err
	}

	return c, err
}

func checkAccess(c *teleport) error {
	if accessNode(c) && accessUser(c) {
		return nil
	}
	//goland:noinspection GoErrorStringFormat
	return errors.New("You do not have access to this server")
}

func accessNode(c *teleport) bool {
	for _, v := range nn {
		if v.Node == c.Node {
			return true
		}
	}
	return false
}

func accessUser(c *teleport) bool {
	for _, l := range s.Logins {
		if l == c.User {
			return true
		}
	}
	return false
}

func setTeleportStatus() error {
	cmdStatus := []string{tsh, "status", "-f", "json"}
	out, err := exec.Command("bash", "-c", strings.Join(cmdStatus[:], " ")).CombinedOutput()
	if err != nil {
		//goland:noinspection GoErrorStringFormat
		return errors.New("The user is not authorized in Teleport")
	}

	err = njson.Unmarshal(out, &s)
	if err != nil {
		return err
	}

	return nil
}

func setTeleportNodes() error {
	cmdLs := []string{tsh, "ls", "-f", "json"}
	out, err := exec.Command("bash", "-c", strings.Join(cmdLs[:], " ")).CombinedOutput()
	if err != nil {
		return err
	}

	var result []json.RawMessage
	err = json.Unmarshal(out, &result)
	if err != nil {
		return err
	}

	var n node
	for _, val := range result {
		err = njson.Unmarshal(val, &n)
		if err != nil {
			return err
		}
		nn = append(nn, n)
	}
	return nil
}

func teleportBin() (string, error) {
	t, err := exec.LookPath("tsh")
	if err != nil {
		return "", err
	}
	return t, nil
}

func (t *teleport) run(cmd string) (string, error) {
	cmdRun := []string{tsh, "ssh", t.User + "@" + t.Node, strconv.Quote(cmd)}
	out, err := exec.Command("bash", "-c", strings.Join(cmdRun[:], " ")).CombinedOutput()
	if err != nil {
		//goland:noinspection GoErrorStringFormat
		return "", errors.New("Something went wrong")
	}

	return string(out), nil
}

func (t *teleport) download(from, to string) error {
	cmdRun := []string{tsh, "scp", "--login=" + t.User, t.Node + ":" + from, to}
	_, err := exec.Command("bash", "-c", strings.Join(cmdRun[:], " ")).CombinedOutput()
	if err != nil {
		//goland:noinspection GoErrorStringFormat
		return errors.New("Something went wrong")
	}

	return nil
}

func (t *teleport) delete(path string) error {
	cmdRun := []string{tsh, "ssh", t.User + "@" + t.Node, strconv.Quote("rm " + path)}
	_, err := exec.Command("bash", "-c", strings.Join(cmdRun[:], " ")).CombinedOutput()
	if err != nil {
		//goland:noinspection GoErrorStringFormat
		return errors.New("Something went wrong")
	}

	return nil
}
