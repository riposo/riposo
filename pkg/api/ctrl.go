package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/riposo/riposo/pkg/conn/cache"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/identity"
	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

type request struct {
	HTTP *http.Request
	Txn  *Txn
	Path riposo.Path
}

func newRequest(req *http.Request) *request {
	return &request{
		HTTP: req,
		Txn:  GetTxn(req),
		Path: GetPath(req),
	}
}

type controller struct {
	act Actions
	cfg *Config
}

func (c *controller) List(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)

	// run through common part
	params, err := c.prepareBulkGet(out, req)
	if err != nil {
		return err
	}

	// paginate objects
	objs, err := c.paginate(out, req, params, "")
	if err != nil {
		return err
	}

	// unauthorized if empty and unauthenticated
	if len(objs) == 0 && req.Txn.User.ID == riposo.Everyone {
		return schema.MissingAuthToken
	}

	return &schema.Objects{Data: objs}
}

func (c *controller) Count(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)

	// run through common part
	params, err := c.prepareBulkGet(out, req)
	if err != nil {
		return err
	}

	// count objects
	count, err := req.Txn.Store.CountAll(req.Path, storage.CountOptions{
		Condition: params.Condition,
	})
	if err != nil {
		return err
	}

	// set extra headers
	total := strconv.FormatInt(count, 10)
	out.Set("Total-Objects", total)
	out.Set("Total-Records", total)
	return nil
}

func (c *controller) DeleteBulk(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)

	// conditional render check
	if err := renderConditional(out, req.HTTP, 0, nil); err != nil {
		return err
	}

	// parse params
	params, err := c.parseBulkQuery(req)
	if err != nil {
		return err
	}

	// if a pagination token is given, ensure it can be 'spent'
	if params.Token != nil {
		if params.Token.Nonce == "" {
			return schema.InvalidQuery("_token has invalid content")
		}

		if err := req.Txn.Cache.Del(params.Token.Nonce); errors.Is(err, cache.ErrNotFound) {
			return schema.InvalidQuery("_token was already used or has expired")
		} else if err != nil {
			return err
		}
	}

	// check if parent exists
	if err := c.checkParent(req); err != nil {
		return err
	}

	// paginate objects
	objs, err := c.paginate(out, req, params, "pagination-token-"+req.Txn.Helpers.NextID())
	if err != nil {
		return err
	}

	// perform action
	modTime, err := c.act.DeleteAll(req.Txn, req.Path, objs)
	if err != nil {
		return err
	}

	// mark as deleted
	for _, obj := range objs {
		obj.Deleted = true
		obj.ModTime = modTime
	}

	// set headers + respond
	setCacheHeaders(out, req.HTTP, modTime)
	return &schema.Objects{Data: objs}
}

func (c *controller) Get(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)
	return c.doGet(out, req)
}

func (c *controller) Create(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)

	// parse payload
	var payload schema.Resource
	if err := Parse(req.HTTP, &payload); err != nil {
		return err
	}

	// ensure data is not nil
	if payload.Data == nil {
		payload.Data = &schema.Object{}
	}

	// validate ID
	if payload.Data.ID != "" && !identity.IsValid(payload.Data.ID) {
		return schema.InvalidPath("Invalid object id")
	}

	// perform create
	return c.createOrGet(out, req, &payload)
}

func (c *controller) Update(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)
	objID := req.Path.ObjectID()

	// validate ID
	if !identity.IsValid(objID) {
		return schema.InvalidPath("Invalid object id")
	}

	// parse payload
	var payload schema.Resource
	if err := Parse(req.HTTP, &payload); err != nil {
		return err
	}

	// ensure requested ID matches body
	if payload.Data == nil {
		payload.Data = &schema.Object{ID: objID}
	} else if payload.Data.ID == "" {
		payload.Data.ID = objID
	} else if payload.Data.ID != objID {
		return schema.InvalidBody("data.id", "Does not match requested object")
	}

	// fetch existing, ignore not-found errors
	exst, err := req.Txn.Store.GetForUpdate(req.Path)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	}

	// render create if record doesn't exist
	if exst == nil {
		req.Path = req.Path.WithObjectID("*")
		return c.createOrGet(out, req, &payload)
	}

	// ensure user has write permission
	if err := c.checkPermission(req.Txn, req.Path, "write"); err != nil {
		return err
	}

	// check if parent exists
	if err := c.checkParent(req); err != nil {
		return err
	}

	// perform conditional render check if exists
	if exst != nil {
		if err := renderConditional(out, req.HTTP, exst.ModTime, exst); err != nil {
			return err
		}
	}

	// update resource & permissions
	res, err := c.act.Update(req.Txn, req.Path, exst, &payload)
	if err != nil {
		return err
	}

	// set headers + respond
	setCacheHeaders(out, req.HTTP, res.Data.ModTime)
	return res
}

func (c *controller) Patch(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)
	objID := req.Path.ObjectID()

	// validate ID
	if !identity.IsValid(objID) {
		return schema.InvalidPath("Invalid object id")
	}

	// parse payload
	var payload schema.Resource
	if err := Parse(req.HTTP, &payload); err != nil {
		return err
	}

	// ensure we have "data" or "permissions"
	if payload.Data == nil && payload.Permissions == nil {
		return schema.InvalidBody("", "Provide at least one of data or permissions")
	}

	// ensure requested ID matches body
	if payload.Data == nil {
		payload.Data = &schema.Object{ID: objID}
	} else if payload.Data.ID == "" {
		payload.Data.ID = objID
	} else if payload.Data.ID != objID {
		return schema.InvalidBody("data.id", "Does not match requested object")
	}

	// ensure user has write permission
	if err := c.checkPermission(req.Txn, req.Path, "write"); err != nil {
		return err
	}

	// check if parent exists
	if err := c.checkParent(req); err != nil {
		return err
	}

	// fetch existing
	exst, err := req.Txn.Store.GetForUpdate(req.Path)
	if errors.Is(err, storage.ErrNotFound) {
		return schema.InvalidResource(req.Path)
	} else if err != nil {
		return err
	}

	// conditional render check
	if err := renderConditional(out, req.HTTP, exst.ModTime, exst); err != nil {
		return err
	}

	// patch resource & permissions
	res, err := c.act.Patch(req.Txn, req.Path, exst, &payload)
	if err != nil {
		return err
	}

	// set headers + respond
	setCacheHeaders(out, req.HTTP, res.Data.ModTime)
	return res
}

func (c *controller) Delete(out http.Header, r *http.Request) interface{} {
	req := newRequest(r)
	objID := req.Path.ObjectID()

	// validate ID
	if !identity.IsValid(objID) {
		return schema.InvalidPath("Invalid object id")
	}

	// ensure user has write permission
	if err := c.checkPermission(req.Txn, req.Path, "write"); err != nil {
		return err
	}

	// check if parent exists
	if err := c.checkParent(req); err != nil {
		return err
	}

	// fetch existing
	exst, err := req.Txn.Store.GetForUpdate(req.Path)
	if errors.Is(err, storage.ErrNotFound) {
		return schema.InvalidResource(req.Path)
	} else if err != nil {
		return err
	}

	// conditional render check
	if err := renderConditional(out, req.HTTP, exst.ModTime, exst); err != nil {
		return err
	}

	// delete resource
	deleted, err := c.act.Delete(req.Txn, req.Path, exst)
	if err != nil {
		return err
	}

	// set headers + respond
	setCacheHeaders(out, req.HTTP, deleted.ModTime)
	return &schema.Resource{Data: deleted}
}

func (c *controller) checkPermission(txn *Txn, path riposo.Path, perms ...string) error {
	ents := poolEntSlice()
	defer ents.Release()

	path.Traverse(func(part riposo.Path) bool {
		for _, perm := range perms {
			ents.Append(perm, part)
		}
		return true
	})
	return c.isForbidden(txn, ents)
}

func (c *controller) isForbidden(txn *Txn, ents *entSlice) error {
	if ok, err := c.cfg.Authz.Verify(txn.Perms, txn.User.Principals, ents.S); err != nil {
		return err
	} else if ok {
		return nil
	}

	if txn.User.ID == riposo.Everyone {
		return schema.MissingAuthToken
	}
	return schema.Forbidden
}

func (c *controller) checkParent(req *request) error {
	if parentPath := req.Path.Parent(); parentPath != "" && parentPath.ResourceName() != "bucket" {
		if ok, err := req.Txn.Store.Exists(parentPath); err != nil {
			return err
		} else if !ok {
			return schema.MissingResource(parentPath.ObjectID(), parentPath.ResourceName())
		}
	}
	return nil
}

var emptyObjects = []*schema.Object{}

func (c *controller) paginate(out http.Header, req *request, params *params.Params, nonce string) ([]*schema.Object, error) {
	// TODO: add back pooling?
	objs, err := req.Txn.Store.ListAll(nil, req.Path, storage.ListOptions{
		Condition:  params.Condition,
		Pagination: params.Token.Conditions(),
		Sort:       params.Sort,
		Limit:      params.Limit + 1,
	})
	if err != nil {
		return nil, err
	}

	if params.Limit > 0 && len(objs) == params.Limit+1 {
		lastObj := objs[params.Limit-1]
		objs = objs[:params.Limit]

		// generate next-page URL and set header
		nurl, err := params.NextPageURL(req.HTTP.URL, nonce, lastObj)
		if err != nil {
			return nil, err
		}
		out.Set("Next-Page", nurl.String())

		// store nonce in cache, if one given
		if nonce != "" {
			if err := req.Txn.Cache.Set(nonce, nil, time.Now().Add(c.cfg.Pagination.TokenValidity)); err != nil {
				return nil, err
			}
		}
	}

	if objs == nil {
		objs = emptyObjects
	}
	return objs, nil
}

func (c *controller) prepareBulkGet(out http.Header, req *request) (*params.Params, error) {
	// obtain modTime
	modTime, err := req.Txn.Store.ModTime(req.Path)
	if err != nil {
		return nil, err
	}

	// conditional render check
	if err := renderConditional(out, req.HTTP, modTime, nil); err != nil {
		return nil, err
	}

	// ensure user has read or write permission to the parent resource
	if parentPath := req.Path.Parent(); parentPath != "" {
		if err := c.checkPermission(req.Txn, parentPath, "read", "write"); err != nil {
			return nil, err
		}
	}

	// check if parent exists
	if err := c.checkParent(req); err != nil {
		return nil, err
	}

	// parse params
	params, err := c.parseBulkQuery(req)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (c *controller) parseBulkQuery(req *request) (*params.Params, error) {
	// parse payload
	if err := req.HTTP.ParseForm(); err != nil {
		return nil, schema.InvalidQuery(err.Error())
	}

	// parse params
	pms, err := params.Parse(req.HTTP.Form, c.cfg.Pagination.MaxLimit)
	if err != nil {
		return nil, schema.InvalidQuery(err.Error())
	}

	// check if user has inherited permissions
	ents := poolEntSlice()
	defer ents.Release()

	req.Path.Parent().Traverse(func(part riposo.Path) bool {
		ents.Append("read", part)
		ents.Append("write", part)
		return true
	})
	if ok, err := c.cfg.Authz.Verify(req.Txn.Perms, req.Txn.User.Principals, ents.S); err != nil {
		return nil, err
	} else if ok {
		return pms, nil
	}

	// otherwise, fetch accessible objects
	ents.Reset()
	ents.Append("read", req.Path)
	ents.Append("write", req.Path)

	accessible := poolPathSlice()
	defer accessible.Release()

	accessible.S, err = req.Txn.Perms.GetAccessiblePaths(accessible.S, req.Txn.User.Principals, ents.S)
	if err != nil {
		return nil, err
	}

	// extract objectIDs from pathMap
	objIDs := make([]schema.Value, 0, len(accessible.S))
	for _, path := range accessible.S {
		objIDs = append(objIDs, schema.StringValue(path.ObjectID()))
	}

	// attach condition to limit results to accessible paths only
	pms.Condition = append(pms.Condition, params.Filter{
		Field:    "id",
		Operator: params.OperatorIN,
		Values:   objIDs,
	})

	return pms, nil
}

func (c *controller) doGet(out http.Header, req *request) interface{} {
	// validate ID
	if objID := req.Path.ObjectID(); !identity.IsValid(objID) {
		return schema.InvalidPath("Invalid object id")
	}

	// ensure user has read or write permission
	if err := c.checkPermission(req.Txn, req.Path, "read", "write"); err != nil {
		return err
	}

	// check if parent exists
	if err := c.checkParent(req); err != nil {
		return err
	}

	// get resource
	exst, err := c.act.Get(req.Txn, req.Path)
	if errors.Is(err, storage.ErrNotFound) {
		return schema.InvalidResource(req.Path)
	} else if err != nil {
		return err
	}

	// conditional render check
	if err := renderConditional(out, req.HTTP, exst.Data.ModTime, exst.Data); err != nil {
		return err
	}

	// return payload
	return exst
}

func (c *controller) createOrGet(out http.Header, req *request, payload *schema.Resource) interface{} {
	// ensure user has create permission
	parentPath := req.Path.Parent()
	ents := poolEntSlice()
	defer ents.Release()

	ents.Append(req.Path.ResourceName()+":create", parentPath)
	parentPath.Traverse(func(part riposo.Path) bool {
		ents.Append("write", part)
		return true
	})
	if err := c.isForbidden(req.Txn, ents); err != nil {
		return err
	}

	// check if parent exists
	if err := c.checkParent(req); err != nil {
		return err
	}

	// create resource, try 'get' if exists
	err := c.act.Create(req.Txn, req.Path, payload)
	if errors.Is(err, storage.ErrObjectExists) && payload.Data.ID != "" {
		req.Path = req.Path.WithObjectID(payload.Data.ID)
		return c.doGet(out, req)
	} else if err != nil {
		return err
	}

	// set headers + respond
	payload.StatusCode = http.StatusCreated
	setCacheHeaders(out, req.HTTP, payload.Data.ModTime)
	return payload
}
