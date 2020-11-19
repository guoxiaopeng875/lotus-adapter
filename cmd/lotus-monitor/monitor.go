package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/guoxiaopeng875/lotus-adapter/api/apitypes"
	"github.com/guoxiaopeng875/lotus-adapter/apiwrapper"
	"gopkg.in/resty.v1"
	"net/http"
	"time"
)

type Processor struct {
	// minerID: lotusAPI
	apis     map[address.Address]*apiwrapper.LotusAPIWrapper
	cli      *resty.Client
	proxyUrl string
}

func NewProcessor(apis map[address.Address]*apiwrapper.LotusAPIWrapper, cli *resty.Client, proxyUrl string) *Processor {
	return &Processor{apis: apis, cli: cli, proxyUrl: proxyUrl}
}

func (p *Processor) PushAll() error {
	var mis []*apitypes.PushedMinerInfo
	for mAddr, apiWrapper := range p.apis {
		mi, err := p.getPushedMinerInfo(mAddr, apiWrapper)
		if err != nil {
			return err
		}
		mis = append(mis, mi)
	}
	if len(mis) == 0 {
		return nil
	}
	return p.do(mis)
}

func (p *Processor) do(body interface{}) error {
	resp, err := p.cli.R().SetBody(body).Post(p.proxyUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return fmt.Errorf("push fail, url:%s, time:%s, respStatus:%d, resBody:%s", p.proxyUrl, time.Now().String(), resp.StatusCode(), string(resp.Body()))
}

func (p *Processor) getPushedMinerInfo(mAddr address.Address, apiWrapper *apiwrapper.LotusAPIWrapper) (*apitypes.PushedMinerInfo, error) {
	pi, err := apiWrapper.MinerProvingInfo(context.Background(), mAddr)
	if err != nil {
		return nil, err
	}
	si, err := apiWrapper.SectorsInfo()
	if err != nil {
		return nil, err
	}
	cai, err := apiWrapper.MinerAssetInfo(context.Background(), mAddr)
	if err != nil {
		return nil, err
	}
	wti, err := apiWrapper.WorkerTaskInfo()
	if err != nil {
		return nil, err
	}
	return &apitypes.PushedMinerInfo{
		MinerAddr:        mAddr,
		ProvingInfo:      pi,
		MinerSectorsInfo: si,
		WorkerTaskState:  wti,
		ClusterAssetInfo: cai,
	}, nil
}