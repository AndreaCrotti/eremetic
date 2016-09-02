package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/klarna/eremetic/config"
	"github.com/klarna/eremetic/types"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandlingHelpers(t *testing.T) {

	status := []types.Status{
		types.Status{
			Status: types.TaskState_TASK_RUNNING,
			Time:   time.Now().Unix(),
		},
	}

	Convey("writeJSON", t, func() {
		Convey("Should respond with a JSON and the appropriate status code", func() {
			var wr = httptest.NewRecorder()

			writeJSON(200, "foo", wr)
			contentType := wr.HeaderMap["Content-Type"][0]
			So(contentType, ShouldEqual, "application/json; charset=UTF-8")
			So(wr.Code, ShouldEqual, http.StatusOK)
		})
	})

	Convey("HandleError", t, func() {
		wr := httptest.NewRecorder()

		Convey("It should return an error status code", func() {
			var err = mockError{
				message: "Error",
			}

			handleError(err, wr, "A test error")

			So(wr.Code, ShouldEqual, 422)
			So(strings.TrimSpace(wr.Body.String()), ShouldEqual, "{\"error\":\"Error\",\"message\":\"A test error\"}")
		})
	})

	Convey("renderHTML", t, func() {
		id := "eremetic-task.1234"

		task := types.EremeticTask{
			TaskCPUs: 0.2,
			TaskMem:  0.5,
			Command:  "test",
			Image:    "test",
			Status:   status,
			ID:       id,
		}

		wr := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/task/eremetic-task.1234", nil)

		renderHTML(wr, r, task, id, &config.Config{Version: "test"})

		body, _ := ioutil.ReadAll(wr.Body)
		So(body, ShouldNotBeEmpty)
		So(string(body), ShouldContainSubstring, "html")
	})

	Convey("makeMap", t, func() {
		task := types.EremeticTask{
			TaskCPUs: 0.2,
			TaskMem:  0.5,
			Command:  "test",
			Image:    "test",
			Status:   status,
			ID:       "eremetic-task.1234",
		}

		data := makeMap(task)
		So(data, ShouldContainKey, "CPU")
		So(data, ShouldContainKey, "Memory")
		So(data, ShouldContainKey, "Status")
		So(data, ShouldContainKey, "ContainerImage")
		So(data, ShouldContainKey, "Command")
		So(data, ShouldContainKey, "TaskID")
	})

	Convey("notFound", t, func() {
		wr := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/task/eremetic-task.1234", nil)
		Convey("text/html", func() {
			r.Header.Add("Accept", "text/html")

			notFound(wr, r)
			b, _ := ioutil.ReadAll(wr.Body)
			body := string(b)
			So(wr.Code, ShouldEqual, http.StatusNotFound)
			So(body, ShouldContainSubstring, "<title>404 Not Found | Eremetic</title>")
		})

		Convey("application/json", func() {
			notFound(wr, r)
			So(wr.Code, ShouldEqual, http.StatusNotFound)
			So(wr.Header().Get("Content-Type"), ShouldContainSubstring, "application/json")
		})
	})

}
