package main

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
)

func main() {
	http.Handle("/exec/", websocket.Handler(ExecContainer))
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Printf("starting on port: 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func ExecContainer(ws *websocket.Conn) {
	container := ws.Request().URL.Path[len("/exec/"):]
	if container == "" {
		ws.Write([]byte("container does not exist"))
		return
	}

	ctx := context.Background()
	cli, _ := client.NewEnvClient()
	cli.UpdateClientVersion("1.24")

	res, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          true,
		Cmd:          []string{"/bin/bash"},
	})
	if err != nil {
		panic(err)
	}

	hijackedResponse, err := cli.ContainerExecAttach(ctx, res.ID, types.ExecConfig{
		Tty:    true,
		Detach: false,
	})
	defer hijackedResponse.Close()

	var receiveStdout chan error

	go func() (err error) {
		if ws != nil {
			io.Copy(ws, hijackedResponse.Reader)
		}
		return err
	}()

	go func() error {
		io.Copy(hijackedResponse.Conn, ws)
		if conn, ok := hijackedResponse.Conn.(interface {
			CloseWrite() error
		}); ok {
			if err := conn.CloseWrite(); err != nil {
			}
		}
		return nil
	}()

	if err := <-receiveStdout; err != nil {
		log.Println(err)
	}

	log.Println("connection established")
	log.Println(ws)
}
