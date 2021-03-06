package api

import (
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/integrations"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/adambaumeister/moxsoar/settings"
	"github.com/adambaumeister/moxsoar/tracker"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

var jwtKey = []byte("FakeKeySon!")

type api struct {
	PackIndex *pack.PackIndex
	RunConfig *pack.RunConfig

	Users map[string]*User

	UserDB     *JSONPasswordDB
	SettingsDB *settings.SettingsDB
}

func enableCors(w *http.ResponseWriter) {
	// This is for development work only,
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func Start(addr string,
	pi *pack.PackIndex,
	rc *pack.RunConfig,
	datadir string,
	staticdir string,
	SSLCertificatePath string,
	SSLKeyPath string) {

	userFile := path.Join(datadir, "users.json")
	jpdb := JSONPasswordDB{
		Path: userFile,
	}
	settingsFile := path.Join(datadir, "settings.json")
	sdb := settings.SettingsDB{
		Path: settingsFile,
	}

	defaultAdminUser := User{
		Credentials: Credentials{
			Username: "admin",
			Password: "admin",
		},
		Name: "Default Administrative User",
	}

	a := api{
		PackIndex: pi,
		Users: map[string]*User{
			"admin": &defaultAdminUser,
		},
		RunConfig:  rc,
		UserDB:     &jpdb,
		SettingsDB: &sdb,
	}
	httpMux := http.NewServeMux()
	s := http.Server{Addr: addr, Handler: httpMux}

	// This
	httpMux.Handle("/", http.FileServer(http.Dir(staticdir)))
	httpMux.HandleFunc("/api/auth", a.auth)
	httpMux.HandleFunc("/api/packs", a.getPacks)
	httpMux.HandleFunc("/api/packs/", a.PackRequest)
	httpMux.HandleFunc("/api/adduser", a.addUser)
	httpMux.HandleFunc("/api/refreshauth", refreshAuth)
	httpMux.HandleFunc("/api/packs/clone", a.clonePack)
	httpMux.HandleFunc("/api/packs/activate", a.setPack)
	httpMux.HandleFunc("/api/packs/status", a.packStatus)
	httpMux.HandleFunc("/api/packs/update", a.updatePack)
	httpMux.HandleFunc("/api/packs/save", a.packSave)
	httpMux.HandleFunc("/api/settings", a.settings)
	httpMux.HandleFunc("/api/settings/variable", a.VariablesRequest)
	httpMux.HandleFunc("/api/settings/test", a.TestTrackerSettings)

	if SSLCertificatePath != "" {
		err := s.ListenAndServeTLS(SSLCertificatePath, SSLKeyPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (a *api) auth(writer http.ResponseWriter, request *http.Request) {
	/*
		Authenticate the API client
	*/
	enableCors(&writer)

	// Get an authentication request message JSON
	var creds Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check the PW matches with what's in the DB
	user, ok := a.Users[creds.Username]
	if !ok {
		writer.WriteHeader(http.StatusUnauthorized)
		r := Error{
			Message: "User does not exist!",
		}
		b := MarshalToJson(r)
		_, _ = writer.Write(b)

		return
	}

	c := Hash{}

	checkHashResult := c.Compare(user.Credentials.Password, creds.Password)
	if checkHashResult != nil {
		// If hash doesn't match, check cleartext
		// This lets us populate the default admin password easier
		if !ok || user.Credentials.Password != creds.Password {
			writer.WriteHeader(http.StatusUnauthorized)
			r := Error{
				Message: "Invalid password.",
			}
			b := MarshalToJson(r)
			_, _ = writer.Write(b)

			return
		}
	}

	expirationTime := time.Now().Add(30 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// This is where we sign the token
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Path:    "/",
		Expires: expirationTime,
	})

	http.SetCookie(writer, &http.Cookie{
		Name:    "username",
		Path:    "/",
		Value:   creds.Username,
		Expires: expirationTime,
	})

	r := LoginMessage{
		Message:  "Logged in!",
		Username: creds.Username,
		Settings: *a.SettingsDB.GetSettings(),
	}

	_ = SendJsonResponse(r, writer)

}

func (a *api) TestTrackerSettings(writer http.ResponseWriter, request *http.Request) {
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	s := a.SettingsDB.GetSettings()
	_, err := tracker.GetElkTracker(s)

	if err == nil {
		r := TrackerStatus{
			Connected: true,
			Message:   fmt.Sprintf("Connected to the elasticsearch server at %v", s.Address),
		}
		_ = SendJsonResponse(r, writer)
		return
	}

	r := TrackerStatus{
		Connected: false,
		Message:   fmt.Sprintf("Could not connect to Elasticsearch at %v, using default request tracker (stdout).", s.Address),
	}
	_ = SendJsonResponse(r, writer)
	return

}

func checkAuth(writer http.ResponseWriter, request *http.Request) (*Claims, *jwt.Token) {
	/*
		Validate the auth ticket is still valid
	*/
	enableCors(&writer)

	c, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			writer.WriteHeader(http.StatusUnauthorized)
			return nil, nil
		}
		// For any other type of error, return a bad request status
		writer.WriteHeader(http.StatusBadRequest)
		return nil, nil
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			writer.WriteHeader(http.StatusUnauthorized)
			return nil, nil
		}
		writer.WriteHeader(http.StatusBadRequest)
		return nil, nil
	}
	if !tkn.Valid {
		writer.WriteHeader(http.StatusUnauthorized)
		return nil, nil
	}

	// Finally, return the welcome message to the user, along with their
	// username given in the token
	return claims, tkn
}

func (a *api) setPack(writer http.ResponseWriter, request *http.Request) {
	// Activate a different pack
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	var ar ActivateRequest
	err := json.NewDecoder(request.Body).Decode(&ar)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	a.RunConfig.Shutdown()
	p, err := a.PackIndex.ActivatePack(ar.PackName)

	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	a.RunConfig = pack.GetRunConfig(p.Path, a.SettingsDB.GetSettings())
	a.RunConfig.RunAll()

	r := ActivateResponse{
		Message: fmt.Sprintf("Activated pack %v", p.Name),
	}
	b := MarshalToJson(r)
	_, _ = writer.Write(b)

}

func (a *api) packStatus(writer http.ResponseWriter, request *http.Request) {
	// Activate a different pack
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}
	var ar ActivateRequest

	err := json.NewDecoder(request.Body).Decode(&ar)
	if err != nil {
		SendError(err, writer, http.StatusBadRequest)
	}

	status, err := a.PackIndex.Status(ar.PackName)
	if err != nil {
		SendError(err, writer, http.StatusBadRequest)
	}
	b := MarshalToJson(status)
	_, _ = writer.Write(b)
}

func (a *api) packSave(writer http.ResponseWriter, request *http.Request) {
	// Activate a different pack
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}
	var cr SaveRequest

	err := json.NewDecoder(request.Body).Decode(&cr)
	if err != nil {
		SendError(err, writer, http.StatusBadRequest)
		return
	}

	err = a.PackIndex.Save(cr.PackName, cr.CommitMessage, cr.Author)
	if err != nil {
		fmt.Printf("Here")
		SendError(err, writer, http.StatusBadRequest)
		return
	}
	b := MarshalToJson(StatusMessage{
		Message: "Saved!",
	})
	_, _ = writer.Write(b)
}

func refreshAuth(writer http.ResponseWriter, request *http.Request) {
	// Refresh the token attached to a user

	// First check the autth is actually valid and error out of it isn't
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	claims := &Claims{}
	newExpTime := time.Now().Add(30 * time.Minute)

	claims.ExpiresAt = newExpTime.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: newExpTime,
	})
}

func (a *api) getPacks(writer http.ResponseWriter, request *http.Request) {
	/*
		Get all the content packs on the system
	*/

	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	packs := a.PackIndex.Packs

	r := GetPacksResponse{
		Packs: packs,
	}

	b := MarshalToJson(r)
	_, _ = writer.Write(b)
}

func (a *api) PackOps(writer http.ResponseWriter, request *http.Request) {
	/*
		Functions related to pack manipulation
	*/

	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	s := strings.Split(request.URL.Path, "/")
	if len(s) < 3 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var err error
	rc := a.RunConfig
	var r interface{}
	// We've asked for something other than just the pack itself
	if len(s) == 5 {
		integrationName := s[4]
		r, err = getIntegration(integrationName, rc)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			r = Error{Message: err.Error()}
			return
		}
	} else if len(s) == 6 {
		integrationName := s[4]
		packId, _ := strconv.Atoi(s[5])
		// need to handle this err
		i := getIntegrationObject(integrationName, rc)
		if i == nil {
			writer.WriteHeader(http.StatusBadRequest)
			r = Error{Message: "Integration not found"}
			return
		}
		for _, m := range i.Routes[packId].Methods {
			fn := m.ResponseFile
			fb, err := ioutil.ReadFile(path.Join(i.PackDir, integrationName, fn))
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				r = Error{Message: err.Error()}
				return
			}

			m.ResponseString = string(fb)

		}

		r = GetRoute{
			Route: i.Routes[packId],
		}

	} else {
		r = GetRunnerResponse{
			RunConfig: rc,
		}
	}

	b := MarshalToJson(r)
	_, err = writer.Write(b)
	if err != nil {
		panic("Failed to write response http")
	}
}

func (a *api) clonePack(writer http.ResponseWriter, request *http.Request) {
	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	cloneReq := CloneRequest{}
	err := json.NewDecoder(request.Body).Decode(&cloneReq)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		r := ErrorMessage("Malformed request.")
		_, _ = writer.Write(r)
		return
	}

	//fmt.Printf("debug %v %v", cloneReq.PackName, cloneReq.Repo)
	clonedPack, err := a.PackIndex.GetOrClone(cloneReq.PackName, cloneReq.Repo)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		r := ErrorMessage(err.Error())
		_, _ = writer.Write(r)
		return
	}

	cr := CloneResponse{
		Message: fmt.Sprintf("Sucessfully cloned pack: %v", clonedPack.Name),
	}

	b := MarshalToJson(cr)
	_, err = writer.Write(b)
	if err != nil {
		panic("Failed to write response http")
	}

}

func (a *api) updatePack(writer http.ResponseWriter, request *http.Request) {
	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	uReq := UpdateRequest{}
	err := json.NewDecoder(request.Body).Decode(&uReq)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		r := ErrorMessage("Malformed request.")
		_, _ = writer.Write(r)
		return
	}

	hashstr, err := a.PackIndex.Update(uReq.PackName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		r := ErrorMessage(err.Error())
		_, _ = writer.Write(r)
		return
	}

	r := StatusMessage{
		Message: *hashstr,
	}
	b := MarshalToJson(r)
	_, err = writer.Write(b)
	if err != nil {
		panic("Failed to write response http")
	}
}

func (a *api) addUser(writer http.ResponseWriter, request *http.Request) {

	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	var creds Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	c := Hash{}
	hpwd, err := c.Generate(creds.Password)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	user := User{
		Credentials: Credentials{
			Username: creds.Username,
			Password: hpwd,
		},
	}

	a.Users[creds.Username] = &user

	err = a.UserDB.Write(a.Users)
	if err != nil {
		fmt.Printf("error writing file: %v", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	r := AddUserMessage{
		Message: fmt.Sprintf("Added user: %v", creds.Username),
	}

	b := MarshalToJson(r)
	_, err = writer.Write(b)
	if err != nil {
		panic("Failed to write response http")
	}

}

func getIntegration(name string, rc *pack.RunConfig) (*GetIntegration, error) {
	ints := rc.GetIntegrations()

	r := GetIntegration{}
	for _, integration := range ints {
		if integration.Name == name {
			r.Routes = integration.Routes
			r.Integration = name
			r.Addr = integration.Addr
			r.Port = strings.Split(r.Addr, ":")[1]
			return &r, nil
		}
	}

	return nil, fmt.Errorf("Integration %v not found", name)
}

func getIntegrationObject(name string, rc *pack.RunConfig) *integrations.BaseIntegration {
	ints := rc.Running

	for _, integration := range ints {
		if integration.Name == name {
			return integration
		}
	}
	return nil
}
