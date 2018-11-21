package gateway

import (
	"context"
	"encoding/base64"
	"fmt"
	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmebqVUQQqQFhg74FtQFszUJo22Vpr3e8qBAkvvV4ho9HH/go-ipfs/core/coreapi/interface"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/mr-tron/base58/base58"
	"github.com/textileio/textile-go/core"
	"github.com/textileio/textile-go/crypto"
)

var log = logging.Logger("tex-gateway")

// Host is the instance used by the daemon
var Host *Gateway

// Gateway is a HTTP API for getting files and links from IPFS
type Gateway struct {
	Node   *core.Textile
	server *http.Server
}

// Start creates a gateway server
func (g *Gateway) Start(addr string) {
	gin.SetMode(gin.ReleaseMode)
	if g.Node != nil {
		gin.DefaultWriter = g.Node.Writer()
	}

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusNoContent)
	})
	router.GET("/favicon.ico", func(c *gin.Context) {
		img, err := base64.StdEncoding.DecodeString(favicon)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		c.Render(200, render.Data{Data: img})
	})

	router.GET("/ipfs/:root", g.gatewayHandler)
	router.GET("/ipfs/:root/*path", g.gatewayHandler)
	router.GET("/ipns/:root", g.profileHandler)
	router.GET("/ipns/:root/*path", g.profileHandler)

	g.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	errc := make(chan error)
	go func() {
		errc <- g.server.ListenAndServe()
		close(errc)
	}()
	go func() {
		for {
			select {
			case err, ok := <-errc:
				if err != nil && err != http.ErrServerClosed {
					log.Errorf("gateway error: %s", err)
				}
				if !ok {
					log.Info("gateway was shutdown")
					return
				}
			}
		}
	}()
	log.Infof("gateway listening at %s", g.server.Addr)
}

// Stop stops the gateway
func (g *Gateway) Stop() error {
	ctx, cancel := context.WithCancel(context.Background())
	if err := g.server.Shutdown(ctx); err != nil {
		log.Errorf("error shutting down gateway: %s", err)
		return err
	}
	cancel()
	return nil
}

// Addr returns the gateway's address
func (g *Gateway) Addr() string {
	return g.server.Addr
}

// gatewayHandler handles gateway http requests
func (g *Gateway) gatewayHandler(c *gin.Context) {
	contentPath := c.Param("root") + c.Param("path")

	data := g.getDataAtPath(c, contentPath)

	// attempt decrypt if key present
	key, exists := c.GetQuery("key")
	if exists {
		keyb, err := base58.Decode(key)
		if err != nil {
			log.Errorf("error decoding key %s: %s", key, err)
			c.Status(404)
			return
		}
		plain, err := crypto.DecryptAES(data, keyb)
		if err != nil {
			log.Errorf("error decrypting %s: %s", contentPath, err)
			c.Status(404)
			return
		}
		c.Render(200, render.Data{Data: plain})
		return
	}

	c.Render(200, render.Data{Data: data})
}

var avatarRx = regexp.MustCompile(`/avatar($|/small$|/large$)`)

// profileHandler handles requests for profile info hosted on ipns
// NOTE: avatar is a magic path, will return data behind link at avatar_uri
func (g *Gateway) profileHandler(c *gin.Context) {
	pathp := c.Param("path")
	if len(pathp) > 0 && pathp[len(pathp)-1] == '/' {
		pathp = pathp[:len(pathp)-1]
	}
	var isAvatar bool
	var avatarSize string

	matches := avatarRx.FindStringSubmatch(pathp)
	if len(matches) == 2 {
		pathp = "/avatar_uri"
		isAvatar = true

		switch matches[1] {
		case "/large":
			avatarSize = "large"
		default:
			avatarSize = "small"
		}
	}

	rootId, err := peer.IDB58Decode(c.Param("root"))
	if err != nil {
		log.Errorf("error decoding root %s: %s", c.Param("root"), err)
		c.Status(404)
		return
	}

	pth, err := g.Node.ResolveProfile(rootId)
	if err != nil {
		log.Errorf("error resolving profile %s: %s", c.Param("root"), err)
		c.Status(404)
		return
	}

	contentPath := pth.String() + pathp
	data := g.getDataAtPath(c, contentPath)

	// if this is an avatar request, fetch and return the linked image
	if isAvatar {
		location := string(data)
		if location == "" {
			fallback, _ := c.GetQuery("fallback")
			if fallback == "true" {
				location = fmt.Sprintf("https://avatars.dicebear.com/v2/identicon/%s.svg", c.Param("root"))
				c.Redirect(307, location)
				return
			} else {
				c.Status(404)
				return
			}
		}

		// old style w/ key
		parsed := strings.Split(location, "?key=")
		if len(parsed) == 2 {
			keyb, err := base58.Decode(parsed[1])
			if err != nil {
				log.Errorf("error decoding key %s: %s", parsed[1], err)
				c.Status(404)
				return
			}

			ciphertext, err := g.Node.DataAtPath(parsed[0])
			if err != nil {
				c.Status(404)
				return
			}

			data, err = crypto.DecryptAES(ciphertext, keyb)
			if err != nil {
				log.Errorf("error decrypting %s: %s", parsed[0], err)
				c.Status(404)
				return
			}

			c.Header("Content-Type", "image/jpeg")

		} else {
			pth := fmt.Sprintf("%s/0/%s/d", location, avatarSize)
			data, err = g.Node.DataAtPath(pth)
			if err != nil {
				c.Status(404)
				return
			}

			var stop int
			if len(data) < 512 {
				stop = len(data)
			} else {
				stop = 512
			}
			media := http.DetectContentType(data[:stop])
			if media != "" {
				c.Header("Content-Type", media)
			}
		}

		c.Header("Cache-Control", "public, max-age=172800") // 2 days
	}

	c.Render(200, render.Data{Data: data})
}

// getDataAtPath get raw data or directory links at path
func (g *Gateway) getDataAtPath(c *gin.Context, pth string) []byte {
	data, err := g.Node.DataAtPath(pth)
	if err != nil {
		if err == iface.ErrIsDir {
			links, err := g.Node.LinksAtPath(pth)
			if err != nil {
				log.Errorf("error getting path %s: %s", pth, err)
				c.Status(404)
				return nil
			}

			var list []string
			for _, link := range links {
				list = append(list, "/"+link.Name)
			}

			c.String(200, "%s", strings.Join(list, "\n"))
			return nil
		}

		log.Errorf("error getting path %s: %s", pth, err)
		c.Status(404)
		return nil
	}
	return data
}