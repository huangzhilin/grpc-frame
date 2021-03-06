package grpc_frame

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type StandardResp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

const (
	proxyFlag = "__succ__"
)

func HttpSuccHandler(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	resp := StandardResp{
		Code: 1,
		Data: p,
		Msg:  "操作成功",
	}
	bs, _ := json.Marshal(&resp)
	return errors.New(proxyFlag + string(bs))
}

func HttpErrorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	raw := err.Error()
	if strings.HasPrefix(raw, proxyFlag) {
		raw = raw[len(proxyFlag):]
		w.Write([]byte(raw))
		return
	}

	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}


	resp := StandardResp{
		Code: 0,
		Data: "",
		Msg:  s.Message(),
	}

	reg := regexp.MustCompile(`^__([\d]+?)__`)
	regList :=reg.FindStringSubmatch(s.Message())
	if len(regList)==2{
		resp.Code ,_ = strconv.Atoi(regList[1])
		resp.Msg = strings.Replace(s.Message(),regList[0],"",1)
	}

	bs, _ := json.Marshal(&resp)
	w.Write(bs)
}
