package main

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"strings"
)

func getProfile(c echo.Context) error {
	self, err := ensureLogin(c)
	if self == nil {
		return err
	}

	channels := []ChannelInfo{}
	err = db.Select(&channels, "SELECT * FROM channel ORDER BY id")
	if err != nil {
		return err
	}

	userName := c.Param("user_name")
	var other User
	err = db.Get(&other, "SELECT * FROM user WHERE name = ?", userName)
	if err == sql.ErrNoRows {
		return echo.ErrNotFound
	}
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "profile", map[string]interface{}{
		"ChannelID":   0,
		"Channels":    channels,
		"User":        self,
		"Other":       other,
		"SelfProfile": self.ID == other.ID,
	})
}

func postProfile(c echo.Context) error {
	self, err := ensureLogin(c)
	if self == nil {
		return err
	}

	avatarName := ""
	var avatarData []byte

	if fh, err := c.FormFile("avatar_icon"); err == http.ErrMissingFile {
		// no file upload
	} else if err != nil {
		return err
	} else {
		dotPos := strings.LastIndexByte(fh.Filename, '.')
		if dotPos < 0 {
			return ErrBadReqeust
		}
		ext := fh.Filename[dotPos:]
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif":
			break
		default:
			return ErrBadReqeust
		}

		file, err := fh.Open()
		if err != nil {
			return err
		}
		avatarData, _ = ioutil.ReadAll(file)
		file.Close()

		if len(avatarData) > avatarMaxBytes {
			return ErrBadReqeust
		}

		avatarName = fmt.Sprintf("%x%s", sha1.Sum(avatarData), ext)
	}

	if avatarName != "" && len(avatarData) > 0 {
		ioutil.WriteFile("/isucon7/webapp/public/icons/"+avatarName, avatarData, 0644)
		_, err = db.Exec("UPDATE user SET avatar_icon = ? WHERE id = ?", avatarName, self.ID)
		if err != nil {
			return err
		}
	}

	if name := c.FormValue("display_name"); name != "" {
		_, err := db.Exec("UPDATE user SET display_name = ? WHERE id = ?", name, self.ID)
		if err != nil {
			return err
		}
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func getIcon(c echo.Context) error {
	var name string
	var data []byte
	err := db.QueryRow("SELECT name, data FROM image WHERE name = ?",
		c.Param("file_name")).Scan(&name, &data)
	if err == sql.ErrNoRows {
		return echo.ErrNotFound
	}
	if err != nil {
		return err
	}

	mime := ""
	switch true {
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"):
		mime = "image/jpeg"
	case strings.HasSuffix(name, ".png"):
		mime = "image/png"
	case strings.HasSuffix(name, ".gif"):
		mime = "image/gif"
	default:
		return echo.ErrNotFound
	}
	return c.Blob(http.StatusOK, mime, data)
}
