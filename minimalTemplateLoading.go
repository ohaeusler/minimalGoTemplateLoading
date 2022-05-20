package web

import (
	"html/template"
	"sync"

	"github.com/SkyPlayerTV/gOliLog"
)

type templateData struct {
	Name string
	Tmpl *template.Template
}

var templateLogger gOliLog.GOliLogger = gOliLog.InitLogger("panel.schule - Template", 1)

type templateLoader struct {
	cacheActive   bool
	path          string
	cache         map[string]*template.Template
	cacheLock     sync.RWMutex
	fileExtension string
	baseTemplates []string
}

//CreateTemplateLoader returns a TemplateLoader, which is used to load templates. It handles locking the cache and adding baseTemplates to each call
//cacheActive activates the template cache
//basePath must be added if the templates are not in the current directory
//fileExtension must be the extension of the template files. The leading dot must be included.
//baseTemplates is optional. It should be a list of template names which should be added to each template to be loaded.
func CreateTemplateLoader(cacheActive bool, basePath string, fileExtension string, baseTemplates ...string) *templateLoader {
	if basePath == "" {
		basePath = "."
	}
	t := templateLoader{
		cacheActive:   cacheActive,
		path:          basePath,
		cache:         make(map[string]*template.Template),
		cacheLock:     sync.RWMutex{},
		fileExtension: fileExtension,
		baseTemplates: []string{},
	}
	for _, name := range baseTemplates {
		t.baseTemplates = append(t.baseTemplates, basePath+"/"+name+fileExtension)
	}
	return &t
}

func (tL *templateLoader) cacheTemplate(tmplData templateData) {
	tL.cacheLock.Lock()
	tL.cache[tmplData.Name] = tmplData.Tmpl
	tL.cacheLock.Unlock()
}

//GetTemplate loads a template. If the template isn't cached, it well be loaded from disk an gets cached. name should be the filename without filetype
func (tL *templateLoader) GetTemplate(name string) (tmpl *template.Template) {
	tL.cacheLock.RLock()
	tmpl, existing := tL.cache[name]
	go tL.cacheLock.RUnlock()
	if !existing || !tL.cacheActive {
		tmpls := append(tL.baseTemplates, tL.path+"/"+name+tL.fileExtension)
		tmpl, _ = template.Must(template.ParseFiles(tmpls...)).Clone()
		tmplData := templateData{
			Name: name,
			Tmpl: tmpl,
		}

		if tL.cacheActive {
			tL.cacheTemplate(tmplData)
			go templateLogger.Log(1, "Cached template: "+name)
		}
	}
	return
}

//asyncGetTemplate is an async wrapper for GetTemplate
func (tL *templateLoader) AsyncGetTemplate(name string) (c chan *template.Template) {
	c = make(chan *template.Template)
	go func() {

		c <- tL.GetTemplate(name)
	}()
	return
}
