package gsession

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

// SessionIDError indicates a problem with reading or decoding a session id
// from an existing cookie.
type SessionIDError struct {
	Cause error
}

func (e *SessionIDError) Error() string {
	return fmt.Sprintf("failed to read session id: %v", e.Cause)
}

// SessionValuesError indicates a problem with retrieving stored values for
// a session.
type SessionValuesError struct {
	Cause error
}

func (e *SessionValuesError) Error() string {
	return fmt.Sprintf("failed to read session values: %v", e.Cause)
}

// Storage provides a backing store for session values
type Storage interface {
	Save(ctx context.Context, id string, values map[interface{}]interface{}) error
	Load(ctx context.Context, id string) (map[interface{}]interface{}, error)
	Delete(ctx context.Context, id string) error
}

// Store implements a generic gorilla sessions.Store. It maintains a session id via
// a http.Cookie while storing the session data via the Storage interface.
type Store struct {
	// Options contains default options for a session. Options dictate the settings
	// for the cookie used to track the session's id.
	Options *sessions.Options

	// IDGenerator should create unique ids for each new session. The id should be
	// suitable for inclusion as a cookie value. If not set an UUID based generator is used.
	IDGenerator func() (string, error)

	// Storage provides a backing store for session values
	Storage Storage

	// CookieCodec can encode session ids before they are written to a client cookie. If not
	// set then session ids are stored in cookies unaltered.
	CookieCodec CookieCodec
}

// Get returns the current session. A cached session is returned if available.
// A new session is created if not already existing.
func (s *Store) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New loads the current session from storage or creates a new one.
func (s *Store) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	if s.Options != nil {
		opts := *s.Options
		session.Options = &opts
	}
	session.IsNew = true

	existingSessionID, err := s.readSessionID(r, name)
	if err != nil {
		return session, &SessionIDError{err}
	}
	if existingSessionID == "" {
		return session, nil
	}

	values, err := s.Storage.Load(r.Context(), existingSessionID)
	if err != nil {
		return session, &SessionValuesError{err}
	}
	session.Values = values
	session.IsNew = false

	return session, nil
}

// Save persists the session to storage and links it to the request via a cookie containing
// the id.
func (s *Store) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Marked for deletion.
	if session.Options.MaxAge < 0 {
		if session.ID != "" {
			if err := s.Storage.Delete(r.Context(), session.ID); err != nil {
				return err
			}
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))

		return nil
	}

	if session.ID == "" {
		id, err := s.generateID()
		if err != nil {
			return fmt.Errorf("idGenerator: %v", err)
		}
		if id == "" {
			return errors.New("generated id was blank")
		}
		session.ID = id
	}

	if err := s.Storage.Save(r.Context(), session.ID, session.Values); err != nil {
		return fmt.Errorf("Storage.Save: %w", err)
	}

	if err := s.writeSessionID(w, session); err != nil {
		return fmt.Errorf("writeSessionID: %w", err)
	}

	return nil
}

func (s *Store) generateID() (string, error) {
	idGenerator := s.IDGenerator
	if idGenerator == nil {
		idGenerator = uuidIDGenerator
	}
	return idGenerator()
}

func (s *Store) readSessionID(r *http.Request, sessionName string) (string, error) {
	c, err := r.Cookie(sessionName)
	if err == http.ErrNoCookie {
		return "", nil
	} else if err != nil {
		return "", err
	}

	if s.CookieCodec == nil {
		return c.Value, nil
	}

	id, err := s.CookieCodec.Decode(sessionName, c.Value)
	if err != nil {
		return "", fmt.Errorf("CookieCodec.Decode: %w", err)
	}

	return id, nil
}

func (s *Store) writeSessionID(w http.ResponseWriter, session *sessions.Session) error {
	id := session.ID
	if s.CookieCodec != nil {
		encoded, err := s.CookieCodec.Encode(session.Name(), id)
		if err != nil {
			return fmt.Errorf("CookieCodec.Encode: %w", err)
		}
		id = encoded
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), id, session.Options))

	return nil
}

func uuidIDGenerator() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("uuid.NewRandom: %v", err)
	}
	return id.String(), nil
}
