package util

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"encoding/json"

	"github.com/golang/glog"
)

func FUNCTION_NAME() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func FUNCTION_NAME_SHORT() string {
	pc, _, _, _ := runtime.Caller(1)
	longName := runtime.FuncForPC(pc).Name()
	return longName[strings.LastIndex(longName, "/")+1:]
}

func CALLER_FUNCTION_NAME() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

func PrintFatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Stdout.Sync()

	glog.Fatalf(format, args...)
}

func LogObjAsJson(level glog.Level, obj interface{}, name string, indent bool) {
	var obj_json []byte
	var err error

	if indent {
		obj_json, err = json.MarshalIndent(obj, "", "  ")
	} else {
		obj_json, err = json.Marshal(obj)
	}

	if err != nil {
		glog.V(level).Infof("%s: %s\n", name, err)
	} else {
		glog.V(level).Infof("%s: %s\n", name, obj_json)
	}
}

/*

package util

import (
    "fmt"
    "io"
    "os"
    "regexp"
    "runtime"
    "strings"

    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"

    "github.com/golang/glog"
)

func FUNCTION_NAME() string {
    pc, _, _, _ := runtime.Caller(1)
    return runtime.FuncForPC(pc).Name()
}

func CALLER_FUNCTION_NAME() string {
    pc, _, _, _ := runtime.Caller(2)
    return runtime.FuncForPC(pc).Name()
}

func PrintFatalf(format string, args ...interface{}) {
    fmt.Printf(format, args...)
    os.Stdout.Sync()

    glog.Fatalf(format, args...)
}

func LogObjAsJson(level glog.Level, obj interface{}, name string, indent bool) {
    var obj_json []byte
    var err error

    if indent {
        obj_json, err = json.MarshalIndent(obj, "", "  ")
    } else {
        obj_json, err = json.Marshal(obj)
    }

    if err != nil {
        glog.V(level).Infof("%s: %s\n", name, err)
    } else {
        glog.V(level).Infof("%s: %s\n", name, obj_json)
    }
}

var EndlineRe = regexp.MustCompile(`\r?\n`)

func ReplaceEndlines(message interface{}, separator string) string {
    var text string

    switch message.(type) {
    case []byte:
        text = string(message.([]byte))
    case string:
        text = message.(string)
    default:
        text = fmt.Sprint(message)
    }

    return EndlineRe.ReplaceAllString(strings.TrimSpace(text), separator)
}

func CopyReceived(dst http.ResponseWriter, src io.Reader, buf []byte) (written int64, err error) {
    defer glog.V(2).Infof("CopyReceived #%d, %+v\n", written, err)

    if buf == nil {
        buf = make([]byte, 32*1024)
    }
    for {
        nr, er := src.Read(buf)
        if er != nil {
            if er != io.EOF {
                err = er
            }
            break
        }
        if nr > 0 {
            nw, ew := dst.Write(buf[0:nr])
            if nw > 0 {
                written += int64(nw)
            }

            if ew != nil {
                err = ew
                break
            }
            if nr != nw {
                err = io.ErrShortWrite
                break
            }

            if f, ok := dst.(http.Flusher); ok {
                f.Flush()
            }
        }
    }

    return written, err
}

func CopyResponse(dst http.ResponseWriter, src io.Reader) (written int64, err error) {
    return CopyReceived(dst, src, nil)
}

func ProxyRequest(method string, reqUrl string, w http.ResponseWriter, toInfo string) {
    glog.V(3).Infoln("Proxy to:", toInfo)
    glog.V(1).Infof("--> curl -X %s '%s'\n", method, reqUrl)

    req, err := http.NewRequest(method, reqUrl, nil)
    if err != nil {
        glog.Warningln("Remote connecting ERR, ", err)
        return
    }
    req.Header.Add("Accept-Encoding", "identity")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        glog.Warningf("Remote %s ERR, %+v\n", method, err)
        return
    }
    defer resp.Body.Close()

    w.WriteHeader(resp.StatusCode)
    nbytes, err := CopyResponse(w, resp.Body)
    glog.V(1).Infof("<-- %s (%+v) #%d\n", resp.Status, err, nbytes)

    if err != nil {
        glog.Warningln("Remote streaming ERR, ", err)
    }
}

func SendRequest(method string, reqUrl string, okStatus int) ([]byte, int, error) {
    glog.V(1).Infof("--> curl -X %s '%s'\n", method, reqUrl)

    emptyBody := []byte{}

    if req, err := http.NewRequest(method, reqUrl, nil); err == nil {
        if resp, err := http.DefaultClient.Do(req); err == nil {
            defer resp.Body.Close()

            body, err := ioutil.ReadAll(resp.Body)
            if err == nil {
                glog.V(2).Infof("<-- %s (%+v) #%d\n", resp.Status, err, len(body))
            } else {
                glog.V(1).Infof("<-- %s (%+v)# %d\n", resp.Status, err, len(body))
            }
            glog.V(3).Infof("%s\n", ReplaceEndlines(body, " "))

            if err == nil && resp.StatusCode != okStatus {
                return body, resp.StatusCode, fmt.Errorf("Response status %d != %d", resp.StatusCode, okStatus)
            }

            return body, resp.StatusCode, err
        } else {
            return emptyBody, http.StatusInternalServerError, err
        }
    } else {
        return emptyBody, 0, err
    }
}

func BuildPathUrl(root string, relPath string) string {
    //checkUrl(root)

    rootUrl, _ := url.ParseRequestURI(root)
    relUrl, _ := url.Parse(relPath)

    return rootUrl.ResolveReference(relUrl).String()
}


*/
