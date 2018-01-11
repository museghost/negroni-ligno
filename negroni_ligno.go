package negroniligno

import (
	"net/http"
	"net/url"
	"time"
	"strconv"

	"github.com/museghost/ligno"
	"github.com/urfave/negroni"
)


type LignoLogger struct {
	log					 *ligno.Logger
	// context map[string]interface{}
	ctx					ligno.Ctx
	// http response
	res					negroni.ResponseWriter

	// variables
	excludes 			[]string
	id					string
	realIP				string
	latency				time.Duration
	bytesIn				string
	startTime			time.Time
}


func InitLignoLogger(logger *ligno.Logger) *LignoLogger {
	return &LignoLogger{
		log: 			logger,
		ctx:			make(ligno.Ctx),
	}
}


func (m *LignoLogger) SetExclude(u string) error {
	if _, err := url.Parse(u); err != nil {
		return err
	}
	m.excludes = append(m.excludes, u)
	return nil
}


func (m *LignoLogger) Excludes() []string {
	return m.excludes
}


func (m *LignoLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	// skip the URL
	for _, u := range m.excludes {
		if r.URL.Path == u {
			next(rw, r)
			return
		}
	}

	// time
	m.startTime = time.Now()

	// call handler
	next(rw, r)

	// after the handler called
	m.res = rw.(negroni.ResponseWriter)
	m.latency = time.Since(m.startTime)

	// id
	if m.id = r.Header.Get("X-Request-Id"); len(m.id) > 0 {
		m.ctx["id"] = m.id
	}

	// real IP
	m.ctx["remote_ip"] = r.RemoteAddr
	if m.realIP = r.Header.Get("X-Real-IP"); len(m.realIP) > 0 {
		m.ctx["remote_ip"] = m.realIP
	}

	// host
	m.ctx["host"] = r.Host

	// uri
	m.ctx["uri"] = r.RequestURI

	// method (POST, GET, etc)
	m.ctx["method"] = r.Method

	// path
	m.ctx["path"] = r.URL.Path
	if len(r.URL.Path) <= 0 {
		m.ctx["path"] = "/"
	}

	// referer
	m.ctx["referer"] = r.Referer()

	// user_agent
	m.ctx["user_agent"] = r.UserAgent()

	// status
	m.ctx["status"] = m.res.Status()

	// text_status
	m.ctx["text_status"] = http.StatusText(m.res.Status())

	// latency
	m.ctx["latency_human"] = m.latency
	m.ctx["latency"] = m.latency.Nanoseconds()

	// bytes_in
	m.bytesIn = r.Header.Get("Content-Length")
	if len(m.bytesIn) <= 0 {
		m.ctx["bytes_in"] = "0"
	} else {
		m.ctx["bytes_in"] = m.bytesIn
	}

	// bytes_out
	m.ctx["bytes_out"] = strconv.FormatInt(int64(m.res.Size()), 10)


	m.log.InfoCtx("request", m.ctx)
}
