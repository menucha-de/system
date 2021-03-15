package journal

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
	"github.com/gorilla/mux"
	"github.com/menucha-de/logging"
)

var log *logging.Logger = logging.GetLogger("journal")

func getJournal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Header().Set("Content-Disposition", "attachment; filename=journal"+fmt.Sprint(time.Now().Unix())+".txt")
	query := r.URL.Query()
	log.Debug("Query ", query)
	params := mux.Vars(r)
	log.Debug("Params ", params)

	conf := sdjournal.JournalReaderConfig{}

	if val, ok := params["unit"]; ok {
		conf.Matches = []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: val + ".service",
			},
		}
	}

	if val, ok := query["lines"]; ok && len(val) > 0 {
		parsed, _ := strconv.Atoi(val[0])
		conf.NumFromTail = uint64(parsed)
	}

	jr, err := sdjournal.NewJournalReader(conf)
	if err != nil {
		log.Error("Can not open journal ", err)
	}

	//jr.Follow(nil, os.Stdout)
	b := make([]byte, 1024) // 1Kb.
	sb := strings.Builder{}
	for {
		c, err := jr.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Info("Error reading ", err)
		}
		sb.WriteString(string(b[:c]))
	}

	_, err = w.Write([]byte(sb.String()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
