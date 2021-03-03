package mount

import (
	"bytes"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
	"github.com/peramic/App.Systemd/service"
	"github.com/peramic/logging"
)

var log *logging.Logger = logging.GetLogger("systemd-mount")

// Service ...
type Service int

// CreateRequest ...
type CreateRequest struct {
	Description string
	What        string
	Where       string
	Options     string
	Lazy        string
}

// Response ...
type Response int

// CreateMount creates a systemd mount unit
func (s *Service) CreateMount(req CreateRequest, resp *Response) error {
	arr := []string{
		"<description>", req.Description,
		"<what>", req.What,
		"<where>", req.Where,
		"<options>", req.Options,
		"<lazy>", req.Lazy,
	}
	// generate filename from given mountpoint (where)
	var cmd *exec.Cmd
	command := "systemd-escape"
	params := []string{"--suffix=mount", "--path", req.Where}
	log.Infof("Executing \"%s %s\"", command, params)
	cmd = exec.Command(command, params...)
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	filename := out.String()
	log.Infof("Mount unit file name: ", out.String())

	service.WriteUnitFile("/etc/systemd/system/"+filename, "mount", arr...)
	return nil
}

func processMount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	if len(name) > 0 {
		var s Service = 0
		req := CreateRequest{"myDescription", "myMount", "my/mount/path/", "myOptions", "false"}
		var resp Response = 0
		err := s.CreateMount(req, &resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
