package cluster

/*
import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/fsouza/go-dockerclient"
	dtesting "github.com/fsouza/go-dockerclient/testing"
	"github.com/megamsys/libgo/safe"
)

func TestRemoveImage(t *testing.T) {
	var called bool
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server1.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	name := "megam/python"
	err = cluster.storage().StoreImage(name, "id1", server1.URL)
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.RemoveImage(name)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("RemoveImage(%q): Did not call node HTTP server", name)
	}
	_, err = cluster.storage().RetrieveImage(name)
	if err != ErrNoSuchImage {
		t.Errorf("RemoveImage(%q): wrong error. Want %#v. Got %#v.", name, ErrNoSuchImage, err)
	}
}

func TestRemoveImageNotFoundInStorage(t *testing.T) {
	var called bool
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server1.Close()
	cluster, err := New(&MapStorage{}, Node{Address: server1.URL})
	if err != nil {
		t.Fatal(err)
	}
	name := "megam/python"
	err = cluster.RemoveImage(name)
	if err != ErrNoSuchImage {
		t.Errorf("RemoveImage(%q): wrong error. Want %#v. Got %#v.", name, ErrNoSuchImage, err)
	}
	if called {
		t.Errorf("RemoveImage(%q): server should not be called.", name)
	}
}

func TestRemoveImageNotFoundInServer(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server1.Close()
	name := "megam/python"
	stor := &MapStorage{}
	err := stor.StoreImage(name, "id1", server1.URL)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := New(stor,
		Node{Address: server1.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.RemoveImage(name)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.storage().RetrieveImage(name)
	if err != ErrNoSuchImage {
		t.Errorf("RemoveImage(%q): wrong error. Want %#v. Got %#v.", name, ErrNoSuchImage, err)
	}
}

func TestRemoveImageServerUnavailable(t *testing.T) {
	addr := "http://invalid-server.nowhere.none"
	name := "megam/python"
	stor := &MapStorage{}
	err := stor.StoreImage(name, "id1", addr)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := New(stor,
		Node{Address: addr},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.RemoveImage(name)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.storage().RetrieveImage(name)
	if err != ErrNoSuchImage {
		t.Errorf("RemoveImage(%q): wrong error. Want %#v. Got %#v.", name, ErrNoSuchImage, err)
	}
}

func TestRemoveImageNodeNotInStorage(t *testing.T) {
	called := false
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server1.Close()
	name := "megam/python"
	stor := &MapStorage{}
	err := stor.StoreImage(name, "id1", server1.URL)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := New(stor)
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.RemoveImage(name)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("RemoveImage(%q): Did not call node HTTP server", name)
	}
	_, err = cluster.storage().RetrieveImage(name)
	if err != ErrNoSuchImage {
		t.Errorf("RemoveImage(%q): wrong error. Want %#v. Got %#v.", name, ErrNoSuchImage, err)
	}
}

func TestPullImage(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/images/megam/python/json" {
			w.Write([]byte(`{"Id": "id1"}`))
		} else {
			w.Write([]byte("Pulling from 1!"))
		}
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/images/megam/python/json" {
			w.Write([]byte(`{"Id": "id1"}`))
		} else {
			w.Write([]byte("Pulling from 2!"))
		}
	}))
	defer server2.Close()
	var buf safe.Buffer
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.PullImageOptions{Repository: "megam/python", OutputStream: &buf}
	err = cluster.PullImage(opts, docker.AuthConfiguration{})
	if err != nil {
		t.Error(err)
	}
	alternatives := []string{
		"Pulling from 1!Pulling from 2!",
		"Pulling from 2!Pulling from 1!",
	}
	if r := buf.String(); r != alternatives[0] && r != alternatives[1] {
		t.Errorf("Wrong output: Want %q. Got %q.", "Pulling from 1!Pulling from 2!", buf.String())
	}
	img, err := cluster.storage().RetrieveImage("megam/python")
	if err != nil {
		t.Fatal(err)
	}
	expected := []ImageHistory{
		{Node: server1.URL, ImageId: "id1"},
		{Node: server2.URL, ImageId: "id1"},
	}
	if !reflect.DeepEqual(img.History, expected) {
		expected[0], expected[1] = expected[1], expected[0]
		if !reflect.DeepEqual(img.History, expected) {
			t.Errorf("Wrong output: Want %q. Got %q.", expected, img)
		}
	}
}

func TestPullImageNotFound(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No such image", http.StatusNotFound)
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No such image", http.StatusNotFound)
	}))
	defer server2.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	opts := docker.PullImageOptions{Repository: "megam/python", OutputStream: &buf}
	err = cluster.PullImage(opts, docker.AuthConfiguration{})
	if err == nil {
		t.Error("PullImage: got unexpected <nil> error")
	}
}

func TestPullImageSpecifyNode(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/images/megam/python/json" {
			w.Write([]byte(`{"Id": "id1"}`))
		} else {
			w.Write([]byte("Pulling from 1!"))
		}
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/images/megam/python/json" {
			w.Write([]byte(`{"Id": "id1"}`))
		} else {
			w.Write([]byte("Pulling from 2!"))
		}
	}))
	defer server2.Close()
	var buf safe.Buffer
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.PullImageOptions{Repository: "megam/python", OutputStream: &buf}
	err = cluster.PullImage(opts, docker.AuthConfiguration{}, server2.URL)
	if err != nil {
		t.Error(err)
	}
	expected := "Pulling from 2!"
	if r := buf.String(); r != expected {
		t.Errorf("Wrong output: Want %q. Got %q.", expected, r)
	}
}

func TestPullImageSpecifyMultipleNodes(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Pulling from 1!"))
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/images/megam/python/json" {
			w.Write([]byte(`{"Id": "id1"}`))
		} else {
			w.Write([]byte("Pulling from 2!"))
		}
	}))
	defer server2.Close()
	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/images/megam/python/json" {
			w.Write([]byte(`{"Id": "id1"}`))
		} else {
			w.Write([]byte("Pulling from 3!"))
		}
	}))
	defer server3.Close()
	var buf safe.Buffer
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
		Node{Address: server3.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.PullImageOptions{Repository: "megam/python", OutputStream: &buf}
	err = cluster.PullImage(opts, docker.AuthConfiguration{}, server2.URL, server3.URL)
	if err != nil {
		t.Error(err)
	}
	alternatives := []string{
		"Pulling from 2!Pulling from 3!",
		"Pulling from 3!Pulling from 2!",
	}
	if r := buf.String(); r != alternatives[0] && r != alternatives[1] {
		t.Errorf("Wrong output: Want %q. Got %q.", "Pulling from 2!Pulling from 3!", r)
	}
}

func TestPushImage(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Pushing to server 1!"))
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Pushing to server 2!"))
	}))
	defer server2.Close()
	var buf safe.Buffer
	stor := &MapStorage{}
	err := stor.StoreImage("megam/ruby", "id1", server1.URL)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := New(stor,
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	var auth docker.AuthConfiguration
	err = cluster.PushImage(docker.PushImageOptions{Name: "megam/ruby", OutputStream: &buf}, auth)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`^Pushing to server \d`)
	if !re.MatchString(buf.String()) {
		t.Errorf("Wrong output: Want %q. Got %q.", "Pushing to server [12]", buf.String())
	}
}

func TestPushImageNotFound(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No such image", http.StatusNotFound)
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No such image", http.StatusNotFound)
	}))
	defer server2.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	var auth docker.AuthConfiguration
	err = cluster.PushImage(docker.PushImageOptions{Name: "megam/python", OutputStream: &buf}, auth)
	if err == nil {
		t.Error("PushImage: got unexpected <nil> error")
	}
}

func TestPushImageWithStorage(t *testing.T) {
	var count int
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pushed"))
	}))
	defer server2.Close()
	stor := MapStorage{}
	err := stor.StoreImage("megam/python", "id1", server2.URL)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := New(&stor,
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	var auth docker.AuthConfiguration
	err = cluster.PushImage(docker.PushImageOptions{Name: "megam/python", OutputStream: &buf}, auth)
	if err != nil {
		t.Error(err)
	}
	if count > 0 {
		t.Error("PushImage with storage: should not send request to all servers, but did.")
	}
}

func TestTagImage(t *testing.T) {
	var call string
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = "server1"
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call = "server2"
	}))
	defer server2.Close()
	stor := &MapStorage{}
	err := stor.StoreImage("megam/ruby", "id1", server1.URL)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := New(stor,
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.TagImageOptions{Repo: "myregistry.com/megam/ruby", Force: true}
	err = cluster.TagImage("megam/ruby", opts)
	if err != nil {
		t.Fatal(err)
	}
	if call != "server1" {
		t.Errorf("Wrong call: Want %q. Got %q.", "server1", call)
	}
	img, err := cluster.storage().RetrieveImage("myregistry.com/megam/ruby")
	if err != nil {
		t.Error(err)
	}
	if img.LastId != "id1" {
		t.Errorf("TagImage: wrong id. Want %q. Got %q.", "id1", img.LastId)
	}
}

func TestTagImageNotFound(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No such image", http.StatusNotFound)
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No such image", http.StatusNotFound)
	}))
	defer server2.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.TagImage("megam/ruby", docker.TagImageOptions{})
	if err == nil {
		t.Error("TagImage: got unexpected <nil> error")
	}
}

func TestImportImage(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("importing from 1"))
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("importing from 2"))
	}))
	defer server2.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	var buf safe.Buffer
	opts := docker.ImportImageOptions{
		Repository:   "megam/python",
		Source:       "http://url.to/tar",
		OutputStream: &buf,
	}
	err = cluster.ImportImage(opts)
	if err != nil {
		t.Error(err)
	}
	re := regexp.MustCompile(`^importing from \d`)
	if !re.MatchString(buf.String()) {
		t.Errorf("Wrong output: Want %q. Got %q.", "importing from [12]", buf.String())
	}
}

func TestImportImageWithAbsentFile(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "file not found", http.StatusNotFound)
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "file not found", http.StatusNotFound)
	}))
	defer server2.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	var buf safe.Buffer
	opts := docker.ImportImageOptions{
		Repository:   "megam/python",
		Source:       "/path/to/tar",
		OutputStream: &buf,
	}
	err = cluster.ImportImage(opts)
	if err == nil {
		t.Error("ImportImage: got unexpected <nil> error")
	}
}

func TestBuildImage(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/images/megam/python/json" {
			w.Write([]byte(`{"Id": "id1"}`))
		}
	}))
	defer server1.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	buildOptions := docker.BuildImageOptions{
		Name:         "megam/python",
		Remote:       "http://localhost/Dockerfile",
		InputStream:  nil,
		OutputStream: &buf,
	}
	err = cluster.BuildImage(buildOptions)
	if err != nil {
		t.Error(err)
	}
	_, err = cluster.storage().RetrieveImage("megam/python")
	if err != nil {
		t.Error(err)
	}
}

func TestBuildImageWithNoNodes(t *testing.T) {
	cluster, err := New(&MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	buildOptions := docker.BuildImageOptions{
		Name:         "megam/python",
		Remote:       "http://localhost/Dockerfile",
		InputStream:  nil,
		OutputStream: &buf,
	}
	err = cluster.BuildImage(buildOptions)
	if err == nil {
		t.Error("Should return an error.")
	}
}

type APIImagesList []docker.APIImages

func (a APIImagesList) Len() int           { return len(a) }
func (a APIImagesList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a APIImagesList) Less(i, j int) bool { return a[i].RepoTags[0] < a[j].RepoTags[0] }

func TestListImages(t *testing.T) {
	server1, err := dtesting.NewServer("127.0.0.1:0", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer server1.Stop()
	server2, err := dtesting.NewServer("127.0.0.1:0", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer server2.Stop()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL()},
		Node{Address: server2.URL()},
	)
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.PullImageOptions{Repository: "megam/python1"}
	err = cluster.PullImage(opts, docker.AuthConfiguration{}, server1.URL())
	if err != nil {
		t.Error(err)
	}
	opts = docker.PullImageOptions{Repository: "megam/python2"}
	err = cluster.PullImage(opts, docker.AuthConfiguration{}, server2.URL())
	if err != nil {
		t.Error(err)
	}
	images, err := cluster.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 2 {
		t.Fatalf("Expected images count to be 2, got: %d", len(images))
	}
	sort.Sort(APIImagesList(images))
	if images[0].RepoTags[0] != "megam/python1" {
		t.Fatalf("Expected images megam/python1, got: %s", images[0].RepoTags[0])
	}
	if images[1].RepoTags[0] != "megam/python2" {
		t.Fatalf("Expected images megam/python2, got: %s", images[0].RepoTags[0])
	}
}

func TestListImagesErrors(t *testing.T) {
	server1, err := dtesting.NewServer("127.0.0.1:0", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer server1.Stop()
	server2, err := dtesting.NewServer("127.0.0.1:0", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer server2.Stop()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL()},
		Node{Address: server2.URL()},
	)
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.PullImageOptions{Repository: "megam/python1"}
	err = cluster.PullImage(opts, docker.AuthConfiguration{}, server1.URL())
	if err != nil {
		t.Error(err)
	}
	opts = docker.PullImageOptions{Repository: "megam/python2"}
	err = cluster.PullImage(opts, docker.AuthConfiguration{}, server2.URL())
	if err != nil {
		t.Error(err)
	}
	server2.PrepareFailure("list-images-error", "/images/json")
	defer server2.ResetFailure("list-images-error")
	_, err = cluster.ListImages(docker.ListImagesOptions{All: true})
	if err == nil {
		t.Fatal("Expected error to exist, got <nil>")
	}
}

func TestInspectImage(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id": "id1"}`))
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id": "id2"}`))
	}))
	defer server2.Close()
	stor := &MapStorage{}
	err := stor.StoreImage("megam/ruby", "id1", server1.URL)
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := New(stor,
		Node{Address: server1.URL},
		Node{Address: server2.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	imgData, err := cluster.InspectImage("megam/ruby")
	if err != nil {
		t.Fatal(err)
	}
	if imgData.ID != "id1" {
		t.Fatalf("Expected image id to be 'id1', got: %s", imgData.ID)
	}
}

func TestInspectImageNotFound(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id": "id1"}`))
	}))
	defer server1.Close()
	cluster, err := New(&MapStorage{},
		Node{Address: server1.URL},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.InspectImage("megam/ruby")
	if err != ErrNoSuchImage {
		t.Fatalf("Expected no such image error, got: %#v", err)
	}
}
*/
