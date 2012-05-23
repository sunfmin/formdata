package formdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type Person struct {
	Name        string
	GivenName   string
	Photo       *multipart.FileHeader
	Resume      *multipart.FileHeader
	Gender      int
	Company     *Company
	Departments []*Department
	Projects    []*Project
	Phones      map[string]string
}

type Company struct {
	Name string
}

type Department struct {
	Id   string
	Name string
}

type Project struct {
	Id      string
	Name    string
	Members []*Person
}

func post(w http.ResponseWriter, r *http.Request) {
	var a *Person
	UnmarshalByPrefix(r, &a, "Person")
	body, _ := json.Marshal(&a)
	if len(a.Projects) != 2 {
		panic(string(body))
	}
	fmt.Fprint(w, string(body))
}

func postwithnames(w http.ResponseWriter, r *http.Request) {
	var a *Person
	UnmarshalByNames(r, &a, []string{"Name", "Gender", "Company.Name"})
	body, _ := json.Marshal(&a)
	fmt.Fprint(w, string(body))
}

func postmultipart(w http.ResponseWriter, r *http.Request) {
	var a *Person
	UnmarshalByPrefix(r, &a, "Person")
	f1, _ := a.Photo.Open()
	f2, _ := a.Resume.Open()

	fc1, _ := ioutil.ReadAll(f1)
	fc2, _ := ioutil.ReadAll(f2)
	fmt.Fprint(w, string(fc1))
	fmt.Fprint(w, string(fc2))
}

func TestParseForm(t *testing.T) {
	http.HandleFunc("/postform", post)
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	data := url.Values{}
	data.Add("Person.Name", "Felix")
	data.Add("Person.Gender", "1")
	data.Add("Person.Company.Name", "The Plant")
	data.Add("Person.Phones.Home", "12121212")
	data.Add("Person.Phones.Company", "12332232")
	data.Add("Person.Departments[0].Id", "1")
	data.Add("Person.Departments[1].Id", "2")
	data.Add("Person.Projects[0].Id", "1")
	data.Add("Person.Projects[1].Id", "2")
	data.Add("Person.Projects[0].Members[1].Name", "Juice")
	data.Add("Person.Projects[0].Members[2].Name", "Felix")

	res, err := http.PostForm(ts.URL+"/postform", data)
	if err != nil {
		panic(err)
	}
	b, _ := ioutil.ReadAll(res.Body)
	var p *Person
	json.Unmarshal(b, &p)
	if p.Name != "Felix" {
		t.Errorf("%+v", string(b))
	}
	if p.Projects[1].Id != "2" {
		t.Errorf("%+v", string(b))
	}
}

func TestParseFormByNames(t *testing.T) {
	http.HandleFunc("/postwithnames", postwithnames)
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	data := url.Values{}
	data.Add("Name", "Felix")
	data.Add("Gender", "1")
	data.Add("Company.Name", "The Plant")
	data.Add("Phones.Home", "12121212")
	data.Add("Phones.Company", "12332232")

	res, err := http.PostForm(ts.URL+"/postwithnames", data)
	if err != nil {
		panic(err)
	}
	b, _ := ioutil.ReadAll(res.Body)
	var p *Person
	json.Unmarshal(b, &p)
	if p == nil || p.Name != "Felix" {
		t.Errorf("%+v", string(b))
	}
	if len(p.Phones) != 0 {
		t.Errorf("%+v", string(b))
	}
}

func TestMultipartParseForm(t *testing.T) {
	http.HandleFunc("/postmultipart", postmultipart)
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/postmultipart", strings.NewReader(multipartContent))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundarySHaDkk90eMKgsVUj")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	b, _ := ioutil.ReadAll(res.Body)
	strb := string(b)
	if !strings.Contains(strb, "the file content a") {
		t.Errorf("%+v", strb)
	}
	if !strings.Contains(strb, "the file content b") {
		t.Errorf("%+v", strb)
	}

}

const multipartContent = `

------WebKitFormBoundarySHaDkk90eMKgsVUj
Content-Disposition: form-data; name="Person.Name"

秦
------WebKitFormBoundarySHaDkk90eMKgsVUj
Content-Disposition: form-data; name="Person.GivenName"

俊滨
------WebKitFormBoundarySHaDkk90eMKgsVUj
Content-Disposition: form-data; name="Person.Photo"; filename="filea.txt"
Content-Type: text/plain

the file content a

------WebKitFormBoundarySHaDkk90eMKgsVUj
Content-Disposition: form-data; name="Person.Resume"; filename="fileb.txt"
Content-Type: text/plain

the file content b

------WebKitFormBoundarySHaDkk90eMKgsVUj--
`
