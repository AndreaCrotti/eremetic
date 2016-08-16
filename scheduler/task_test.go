package scheduler

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTask(t *testing.T) {

	status := []types.Status{
		types.Status{
			Status: mesos.TaskState_TASK_RUNNING.String(),
			Time:   time.Now().Unix(),
		},
	}

	Convey("createTaskInfo", t, func() {
		eremeticTask := types.EremeticTask{
			TaskCPUs: 0.2,
			TaskMem:  0.5,
			Command:  "echo hello",
			Image:    "busybox",
			Status:   status,
			ID:       "eremetic-task.1234",
			Name:     "Eremetic task 17",
		}

		offer := mesos.Offer{
			FrameworkId: &mesos.FrameworkID{
				Value: proto.String("framework-id"),
			},
			SlaveId: &mesos.SlaveID{
				Value: proto.String("slave-id"),
			},
			Hostname: proto.String("hostname"),
		}

		Convey("No volume or environment specified", func() {
			net, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.GetName(), ShouldEqual, eremeticTask.Name)
			So(taskInfo.GetResources()[0].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskCPUs)
			So(taskInfo.GetResources()[1].GetScalar().GetValue(), ShouldEqual, eremeticTask.TaskMem)
			So(taskInfo.Container.GetType().String(), ShouldEqual, "DOCKER")
			So(taskInfo.Container.Docker.GetImage(), ShouldEqual, "busybox")
			So(net.SlaveId, ShouldEqual, "slave-id")
			So(taskInfo.Container.Docker.GetForcePullImage(), ShouldBeFalse)
		})

		Convey("Given no Command", func() {
			eremeticTask.Command = ""

			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.Command.GetValue(), ShouldBeEmpty)
			So(taskInfo.Command.GetShell(), ShouldBeFalse)
		})

		Convey("Given a volume and environment", func() {
			volumes := []types.Volume{types.Volume{
				ContainerPath: "/var/www",
				HostPath:      "/var/www",
			}}

			environment := make(map[string]string)
			environment["foo"] = "bar"

			eremeticTask.Environment = environment
			eremeticTask.Volumes = volumes

			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Container.Volumes[0].GetContainerPath(), ShouldEqual, volumes[0].ContainerPath)
			So(taskInfo.Container.Volumes[0].GetHostPath(), ShouldEqual, volumes[0].HostPath)
			So(taskInfo.Command.Environment.Variables[0].GetName(), ShouldEqual, "foo")
			So(taskInfo.Command.Environment.Variables[0].GetValue(), ShouldEqual, "bar")
			So(taskInfo.Command.Environment.Variables[1].GetName(), ShouldEqual, "MESOS_TASK_ID")
			So(taskInfo.Command.Environment.Variables[1].GetValue(), ShouldEqual, eremeticTask.ID)
		})

		Convey("Given archive to fetch", func() {
			URI := []types.URI{types.URI{
				URI:     "http://foobar.local/cats.zip",
				Extract: true,
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeTrue)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeFalse)
		})

		Convey("Given archive to fetch and cache", func() {
			URI := []types.URI{types.URI{
				URI:     "http://foobar.local/cats.zip",
				Extract: true,
				Cache:   true,
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeTrue)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeTrue)
		})

		Convey("Given image to fetch", func() {
			URI := []types.URI{types.URI{
				URI: "http://foobar.local/cats.jpeg",
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeFalse)
		})

		Convey("Given script to fetch", func() {
			URI := []types.URI{types.URI{
				URI:        "http://foobar.local/cats.sh",
				Executable: true,
			}}
			eremeticTask.FetchURIs = URI
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Command.Uris, ShouldHaveLength, 1)
			So(taskInfo.Command.Uris[0].GetValue(), ShouldEqual, eremeticTask.FetchURIs[0].URI)
			So(taskInfo.Command.Uris[0].GetExecutable(), ShouldBeTrue)
			So(taskInfo.Command.Uris[0].GetExtract(), ShouldBeFalse)
			So(taskInfo.Command.Uris[0].GetCache(), ShouldBeFalse)
		})

		Convey("Force pull of docker image", func() {
			eremeticTask.ForcePullImage = true
			_, taskInfo := createTaskInfo(eremeticTask, &offer)

			So(taskInfo.TaskId.GetValue(), ShouldEqual, eremeticTask.ID)
			So(taskInfo.Container.Docker.GetForcePullImage(), ShouldBeTrue)
		})
	})
}
