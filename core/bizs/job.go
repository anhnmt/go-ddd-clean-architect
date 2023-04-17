package bizs

import (
	s "app/core/interfaces/bizs"
	c "app/core/interfaces/caching"
	r "app/core/interfaces/repositories"
	mds "app/core/models"
	"app/pkg/logger"
	"context"
	"net/http"

	"github.com/jinzhu/copier"
)

type JobBiz struct {
	jc     c.JobCache
	jr     r.JobRepo
	logger logger.Logger
}

func NewJob(jobCache c.JobCache, jobRepo r.JobRepo, logger *logger.ApiLogger) s.JobBiz {
	return &JobBiz{
		jc:     jobCache,
		jr:     jobRepo,
		logger: logger,
	}
}

// AddJob implements bizs.JobBiz
func (jb *JobBiz) AddJob(ctx context.Context, job interface{}) (mds.BaseReponse[mds.Job], error) {
	panic("unimplemented")
}

// DeleteJob implements bizs.JobBiz
func (jb *JobBiz) DeleteJob(ctx context.Context, jobId string) (mds.BaseReponse[mds.Job], error) {
	panic("unimplemented")
}

// GetJob implements bizs.JobBiz
func (jb *JobBiz) GetJob(ctx context.Context, jobId string) (mds.BaseReponse[mds.Job], error) {
	panic("unimplemented")
}

// PatchJob implements bizs.JobBiz
func (jb *JobBiz) PatchJob(ctx context.Context, job interface{}) (mds.BaseReponse[mds.Job], error) {
	panic("unimplemented")
}

// SearchJob implements bizs.JobBiz
func (jb *JobBiz) SearchJob(ctx context.Context, index, term string, pageIndex int, pageAmount int) (mds.BaseReponse[mds.Job], error) {
	data, err := jb.jc.FindSimilar(ctx, index, term, pageIndex, pageAmount)
	if err != nil {
		return mds.BaseReponse[mds.Job]{}, err
	}
	return mds.BaseReponse[mds.Job]{
		Code:    http.StatusOK,
		Message: "",
		Errors:  []mds.ResponseError{},
		Data:    data,
	}, nil
}

// SearchJob implements bizs.JobBiz
func (jb *JobBiz) SearchJobDatabase(ctx context.Context, term string, pageIndex int, pageAmount int) (mds.BaseReponse[mds.Job], error) {
	data, err := jb.jr.SearchJob(ctx, term, pageIndex, pageAmount)
	jobs := make([]mds.Job, 0)
	copier.Copy(&jobs, &data)
	if err != nil {
		return mds.BaseReponse[mds.Job]{}, err
	}
	return mds.BaseReponse[mds.Job]{
		Code:    http.StatusOK,
		Message: "",
		Errors:  []mds.ResponseError{},
		Data:    jobs,
	}, nil
}

// UpdateJob implements bizs.JobBiz
func (jb *JobBiz) UpdateJob(ctx context.Context, job interface{}) (mds.BaseReponse[mds.Job], error) {
	panic("unimplemented")
}

// UpdateJob implements bizs.JobBiz
func (jb *JobBiz) CreateIndex(ctx context.Context, index string) error {
	// Define index mapping.
	jb.jc.CreateIndex(ctx, index)

	return nil
}

// UpdateJob implements bizs.JobBiz
func (jb *JobBiz) PushDocuments(ctx context.Context, index string, takes, skips int) (success, fail int, err error) {
	//step 1. query all data from database
	pageIndex := skips / 10000
	pageAmount := 10000
	round := 0

	jb.logger.Infof("PushDocuments index: %v, takes: %v, skips: %v, pageIndex: %v \n", index, takes, skips, pageIndex)
	for takes > 0 {
		data, err := jb.jr.SearchJob(ctx, "", pageIndex, pageAmount)
		if err != nil {
			return 0, 0, err
		}

		jobs := make([]mds.Job, 0)
		if err = copier.Copy(&jobs, &data); err != nil {
			return 0, 0, err
		}

		if len(jobs) == 0 {
			break // there is no more records
		}

		//step 2. bulk index to opensearch
		ss, ff, err := jb.jc.PushDocuments(ctx, index, jobs)
		success += ss
		fail += ff
		pageIndex++
		round++
		takes = takes - pageAmount*round
		if err != nil {
			return success, fail, err
		}
	}

	return success, fail, err
}
