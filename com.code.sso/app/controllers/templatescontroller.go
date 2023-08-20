package controllers

import (
	"html/template"
	"log"
	"net/http"

	"com.code.sso/com.code.sso/config"
	"com.code.sso/com.code.sso/database/basefunctions"
	"com.code.sso/com.code.sso/database/basetypes"
	"com.code.sso/com.code.sso/httpHandler/basecontrollers/baseinterfaces"
	"com.code.sso/com.code.sso/httpHandler/baserouter"
	"com.code.sso/com.code.sso/httpHandler/basevalidators"
)

type TemplatesController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u TemplatesController) GetDBName() basetypes.DBName {
	return basetypes.DBName(config.GetInstance().Database.DBName)
}

func (u TemplatesController) GetCollectionName() basetypes.CollectionName {
	return "templates"
}

func (u TemplatesController) DoIndexing() error {
	return nil
}

func (u *TemplatesController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *TemplatesController) handleSignup(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("com.code.sso/app/view/signup.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (u *TemplatesController) handleSignupEmail(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("com.code.sso/app/view/signupemail.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (u *TemplatesController) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("com.code.sso/app/view/updateprofile.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (u *TemplatesController) handleDashboard(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("com.code.sso/app/view/dashboard.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (u *TemplatesController) handleLogin(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("com.code.sso/app/view/login.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (u TemplatesController) RegisterApis() {
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/signup", u.handleSignup).Methods("GET")
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/signupemail", u.handleSignupEmail).Methods("GET")
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/updateprofile", u.handleUpdateProfile).Methods("GET")
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/dashboard", u.handleDashboard).Methods("GET")
	baserouter.GetInstance().GetBaseRouter().HandleFunc("/login", u.handleLogin).Methods("GET")
}
