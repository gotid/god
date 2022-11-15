package logic

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/gotid/god/examples/upload/internal/svc"
	"github.com/gotid/god/examples/upload/internal/types"

	"github.com/gotid/god/lib/logx"
)

const maxFileSize = 10 << 20 // 10 MB

type UploadLogic struct {
	logx.Logger
	r      *http.Request
	svcCtx *svc.ServiceContext
}

func NewUploadLogic(r *http.Request, svcCtx *svc.ServiceContext) *UploadLogic {
	return &UploadLogic{
		Logger: logx.WithContext(r.Context()),
		r:      r,
		svcCtx: svcCtx,
	}
}

func (l *UploadLogic) Upload() (resp *types.Response, err error) {
	l.r.ParseMultipartForm(maxFileSize)
	file, handler, err := l.r.FormFile("myFile")
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	defer file.Close()

	fmt.Printf("上传文件：%+v\n", handler.Filename)
	fmt.Printf("文件大小：%+v\n", handler.Size)
	fmt.Printf("MIME 头：%+v\n", handler.Header)

	tempFile, err := os.Create(path.Join(l.svcCtx.Config.Path, handler.Filename))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer tempFile.Close()
	io.Copy(tempFile, file)

	return &types.Response{
		Ok: 0,
	}, nil
}
