// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_vtable || vtable
// +build sqlite_vtable vtable

package sqlite3

/*
#cgo CFLAGS: -std=gnu99
#cgo CFLAGS: -DSQLITE_ENABLE_RTREE
#cgo CFLAGS: -DSQLITE_THREADSAFE
#cgo CFLAGS: -DSQLITE_ENABLE_FTS3
#cgo CFLAGS: -DSQLITE_ENABLE_FTS3_PARENTHESIS
#cgo CFLAGS: -DSQLITE_ENABLE_FTS4_UNICODE61
#cgo CFLAGS: -DSQLITE_TRACE_SIZE_LIMIT=15
#cgo CFLAGS: -DSQLITE_ENABLE_COLUMN_METADATA=1
#cgo CFLAGS: -Wno-deprecated-declarations

#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h"
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>
#include <stdint.h>
#include <memory.h>

static inline char *_sqlite3_mprintf(char *zFormat, char *arg) {
  return sqlite3_mprintf(zFormat, arg);
}

typedef struct goVTab goVTab;

struct goVTab {
	sqlite3_vtab base;
	void *vTab;
};

uintptr_t goMInit(void *db, void *pAux, int argc, char **argv, char **pzErr, int isCreate);

static int cXInit(sqlite3 *db, void *pAux, int argc, const char *const*argv, sqlite3_vtab **ppVTab, char **pzErr, int isCreate) {
	void *vTab = (void *)goMInit(db, pAux, argc, (char**)argv, pzErr, isCreate);
	if (!vTab || *pzErr) {
		return SQLITE_ERROR;
	}
	goVTab *pvTab = (goVTab *)sqlite3_malloc(sizeof(goVTab));
	if (!pvTab) {
		*pzErr = sqlite3_mprintf("%s", "Out of memory");
		return SQLITE_NOMEM;
	}
	memset(pvTab, 0, sizeof(goVTab));
	pvTab->vTab = vTab;

	*ppVTab = (sqlite3_vtab *)pvTab;
	*pzErr = 0;
	return SQLITE_OK;
}

static inline int cXCreate(sqlite3 *db, void *pAux, int argc, const char *const*argv, sqlite3_vtab **ppVTab, char **pzErr) {
	return cXInit(db, pAux, argc, argv, ppVTab, pzErr, 1);
}
static inline int cXConnect(sqlite3 *db, void *pAux, int argc, const char *const*argv, sqlite3_vtab **ppVTab, char **pzErr) {
	return cXInit(db, pAux, argc, argv, ppVTab, pzErr, 0);
}

int goVBestIndex(void *pVTab, void *icp, char **pzErr);

static inline int cXBestIndex(sqlite3_vtab *pVTab, sqlite3_index_info *info) {
	return goVBestIndex(((goVTab*)pVTab)->vTab, info, &(pVTab->zErrMsg));
}

char* goVRelease(void *pVTab, int isDestroy);

static int cXRelease(sqlite3_vtab *pVTab, int isDestroy) {
	char *pzErr = goVRelease(((goVTab*)pVTab)->vTab, isDestroy);
	if (pzErr) {
		if (pVTab->zErrMsg)
			sqlite3_free(pVTab->zErrMsg);
		pVTab->zErrMsg = pzErr;
		return SQLITE_ERROR;
	}
	if (pVTab->zErrMsg)
		sqlite3_free(pVTab->zErrMsg);
	sqlite3_free(pVTab);
	return SQLITE_OK;
}

static inline int cXDisconnect(sqlite3_vtab *pVTab) {
	return cXRelease(pVTab, 0);
}
static inline int cXDestroy(sqlite3_vtab *pVTab) {
	return cXRelease(pVTab, 1);
}

typedef struct goVTabCursor goVTabCursor;

struct goVTabCursor {
	sqlite3_vtab_cursor base;
	void *vTabCursor;
};

uintptr_t goVOpen(void *pVTab, char **pzErr);

static int cXOpen(sqlite3_vtab *pVTab, sqlite3_vtab_cursor **ppCursor) {
	void *vTabCursor = (void *)goVOpen(((goVTab*)pVTab)->vTab, &(pVTab->zErrMsg));
	if (!vTabCursor) {
		return SQLITE_ERROR;
	}
	goVTabCursor *pCursor = (goVTabCursor *)sqlite3_malloc(sizeof(goVTabCursor));
	if (!pCursor) {
		return SQLITE_NOMEM;
	}
	memset(pCursor, 0, sizeof(goVTabCursor));
	pCursor->vTabCursor = vTabCursor;
	*ppCursor = (sqlite3_vtab_cursor *)pCursor;
	return SQLITE_OK;
}

static int setErrMsg(sqlite3_vtab_cursor *pCursor, char *pzErr) {
	if (pCursor->pVtab->zErrMsg)
		sqlite3_free(pCursor->pVtab->zErrMsg);
	pCursor->pVtab->zErrMsg = pzErr;
	return SQLITE_ERROR;
}

char* goVClose(void *pCursor);

static int cXClose(sqlite3_vtab_cursor *pCursor) {
	char *pzErr = goVClose(((goVTabCursor*)pCursor)->vTabCursor);
	if (pzErr) {
		return setErrMsg(pCursor, pzErr);
	}
	sqlite3_free(pCursor);
	return SQLITE_OK;
}

char* goVFilter(void *pCursor, int idxNum, char* idxName, int argc, sqlite3_value **argv);

static int cXFilter(sqlite3_vtab_cursor *pCursor, int idxNum, const char *idxStr, int argc, sqlite3_value **argv) {
	char *pzErr = goVFilter(((goVTabCursor*)pCursor)->vTabCursor, idxNum, (char*)idxStr, argc, argv);
	if (pzErr) {
		return setErrMsg(pCursor, pzErr);
	}
	return SQLITE_OK;
}

char* goVNext(void *pCursor);

static int cXNext(sqlite3_vtab_cursor *pCursor) {
	char *pzErr = goVNext(((goVTabCursor*)pCursor)->vTabCursor);
	if (pzErr) {
		return setErrMsg(pCursor, pzErr);
	}
	return SQLITE_OK;
}

int goVEof(void *pCursor);

static inline int cXEof(sqlite3_vtab_cursor *pCursor) {
	return goVEof(((goVTabCursor*)pCursor)->vTabCursor);
}

char* goVColumn(void *pCursor, void *cp, int col, int nochange);

static int cXColumn(sqlite3_vtab_cursor *pCursor, sqlite3_context *ctx, int i) {
	// Check if int sqlite3_vtab_nochange(sqlite3_context*) returns 1
	// If it does, warn goVColumn that the value has not changed
	int nochange = sqlite3_vtab_nochange(ctx);

	char *pzErr = goVColumn(((goVTabCursor*)pCursor)->vTabCursor, ctx, i, nochange);

	if (pzErr) {
		return setErrMsg(pCursor, pzErr);
	}
	return SQLITE_OK;
}

char* goVRowid(void *pCursor, sqlite3_int64 *pRowid);

static int cXRowid(sqlite3_vtab_cursor *pCursor, sqlite3_int64 *pRowid) {
	char *pzErr = goVRowid(((goVTabCursor*)pCursor)->vTabCursor, pRowid);
	if (pzErr) {
		return setErrMsg(pCursor, pzErr);
	}
	return SQLITE_OK;
}

char* goVUpdate(void *pVTab, int argc, sqlite3_value **argv, sqlite3_int64 *pRowid);

int sqlite3_vtab_nochange(sqlite3_context*);
int sqlite3_value_nochange(sqlite3_value*);

static int cXUpdate(sqlite3_vtab *pVTab, int argc, sqlite3_value **argv, sqlite3_int64 *pRowid) {
	char *pzErr = goVUpdate(((goVTab*)pVTab)->vTab, argc, argv, pRowid);
	if (pzErr) {
		if (pVTab->zErrMsg)
			sqlite3_free(pVTab->zErrMsg);
		pVTab->zErrMsg = pzErr;
		return SQLITE_ERROR;
	}
	return SQLITE_OK;
}

char * goVBegin(void *pVTab);

static int cXBegin(sqlite3_vtab *pVTab) {
	char *pzErr = goVBegin(((goVTab*)pVTab)->vTab);
	if (pzErr) {
		if (pVTab->zErrMsg)
			sqlite3_free(pVTab->zErrMsg);
		pVTab->zErrMsg = pzErr;
		return SQLITE_ERROR;
	}
	return SQLITE_OK;
}

char * goVCommit(void *pVTab);

static int cXCommit(sqlite3_vtab *pVTab) {
	char *pzErr = goVCommit(((goVTab*)pVTab)->vTab);
	if (pzErr) {
		if (pVTab->zErrMsg)
			sqlite3_free(pVTab->zErrMsg);
		pVTab->zErrMsg = pzErr;
		return SQLITE_ERROR;
	}
	return SQLITE_OK;
}

char * goVRollback(void *pVTab);

static int cXRollback(sqlite3_vtab *pVTab) {
	char *pzErr = goVRollback(((goVTab*)pVTab)->vTab);
	if (pzErr) {
		if (pVTab->zErrMsg)
			sqlite3_free(pVTab->zErrMsg);
		pVTab->zErrMsg = pzErr;
		return SQLITE_ERROR;
	}
	return SQLITE_OK;
}

static sqlite3_module goModule = {
	0,                       // iVersion
	cXCreate,                // xCreate - create a table
	cXConnect,               // xConnect - connect to an existing table
	cXBestIndex,             // xBestIndex - Determine search strategy
	cXDisconnect,            // xDisconnect - Disconnect from a table
	cXDestroy,               // xDestroy - Drop a table
	cXOpen,                  // xOpen - open a cursor
	cXClose,                 // xClose - close a cursor
	cXFilter,                // xFilter - configure scan constraints
	cXNext,                  // xNext - advance a cursor
	cXEof,                   // xEof
	cXColumn,                // xColumn - read data
	cXRowid,                 // xRowid - read data
	cXUpdate,                // xUpdate - write data
// Not implemented
	0,                       // xBegin - begin transaction
	0,                       // xSync - sync transaction
	0,                       // xCommit - commit transaction
	0,                       // xRollback - rollback transaction
	0,                       // xFindFunction - function overloading
	0,                       // xRename - rename the table
	0,                       // xSavepoint
	0,                       // xRelease
	0	                     // xRollbackTo
};

// See https://sqlite.org/vtab.html#eponymous_only_virtual_tables
static sqlite3_module goModuleEponymousOnly = {
	0,                       // iVersion
	0,                       // xCreate - create a table, which here is null
	cXConnect,               // xConnect - connect to an existing table
	cXBestIndex,             // xBestIndex - Determine search strategy
	cXDisconnect,            // xDisconnect - Disconnect from a table
	cXDestroy,               // xDestroy - Drop a table
	cXOpen,                  // xOpen - open a cursor
	cXClose,                 // xClose - close a cursor
	cXFilter,                // xFilter - configure scan constraints
	cXNext,                  // xNext - advance a cursor
	cXEof,                   // xEof
	cXColumn,                // xColumn - read data
	cXRowid,                 // xRowid - read data
	cXUpdate,                // xUpdate - write data
// Not implemented
	0,                       // xBegin - begin transaction
	0,                       // xSync - sync transaction
	0,                       // xCommit - commit transaction
	0,                       // xRollback - rollback transaction
	0,                       // xFindFunction - function overloading
	0,                       // xRename - rename the table
	0,                       // xSavepoint
	0,                       // xRelease
	0	                     // xRollbackTo
};

static sqlite3_module goModuleTransaction = {
	0,                       // iVersion
	cXCreate,                // xCreate - create a table
	cXConnect,               // xConnect - connect to an existing table
	cXBestIndex,             // xBestIndex - Determine search strategy
	cXDisconnect,            // xDisconnect - Disconnect from a table
	cXDestroy,               // xDestroy - Drop a table
	cXOpen,                  // xOpen - open a cursor
	cXClose,                 // xClose - close a cursor
	cXFilter,                // xFilter - configure scan constraints
	cXNext,                  // xNext - advance a cursor
	cXEof,                   // xEof
	cXColumn,                // xColumn - read data
	cXRowid,                 // xRowid - read data
	cXUpdate,                // xUpdate - write data
	cXBegin,                 // xBegin - begin transaction
	0,                       // xSync - sync transaction
	cXCommit,                // xCommit - commit transaction
	cXRollback,              // xRollback - rollback transaction
	0,                       // xFindFunction - function overloading
	0,                       // xRename - rename the table
	0,                       // xSavepoint
	0,                       // xRelease
	0	                     // xRollbackTo
};

void goMDestroy(void*);

static int _sqlite3_create_module(sqlite3 *db, const char *zName, uintptr_t pClientData) {
  return sqlite3_create_module_v2(db, zName, &goModule, (void*) pClientData, goMDestroy);
}

static int _sqlite3_create_module_eponymous_only(sqlite3 *db, const char *zName, uintptr_t pClientData) {
  return sqlite3_create_module_v2(db, zName, &goModuleEponymousOnly, (void*) pClientData, goMDestroy);
}

static int _sqlite3_create_module_transaction(sqlite3 *db, const char *zName, uintptr_t pClientData) {
  return sqlite3_create_module_v2(db, zName, &goModuleTransaction, (void*) pClientData, goMDestroy);
}

static int _sqlite3_drop_modules(sqlite3 *db, const char **azKeep) {
  return sqlite3_drop_modules(db, azKeep);
}

*/
import "C"

import (
	"database/sql/driver"
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"
)

type sqliteModule struct {
	c      *SQLiteConn
	name   string
	module Module
}

type sqliteVTab struct {
	module        *sqliteModule
	vTab          VTab
	partialUpdate bool
}

type sqliteVTabCursor struct {
	vTab          *sqliteVTab
	vTabCursor    VTabCursor
	partialUpdate bool
}

// Op is type of operations.
type Op uint8

// Op mean identity of operations.
const (
	OpEQ     Op = 2
	OpGT        = 4
	OpLE        = 8
	OpLT        = 16
	OpGE        = 32
	OpMATCH     = 64
	OpLIKE      = 65 /* 3.10.0 and later only */
	OpGLOB      = 66 /* 3.10.0 and later only */
	OpREGEXP    = 67 /* 3.10.0 and later only */
	OpLIMIT     = 73
	OpOFFSET    = 74
)

// InfoConstraint give information of constraint.
type InfoConstraint struct {
	Column int
	Op     Op
	Usable bool
}

// InfoOrderBy give information of order-by.
type InfoOrderBy struct {
	Column int
	Desc   bool
}

func constraints(info *C.sqlite3_index_info) []InfoConstraint {
	slice := *(*[]C.struct_sqlite3_index_constraint)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(info.aConstraint)),
		Len:  int(info.nConstraint),
		Cap:  int(info.nConstraint),
	}))

	cst := make([]InfoConstraint, 0, len(slice))
	for _, c := range slice {
		var usable bool
		if c.usable > 0 {
			usable = true
		}
		cst = append(cst, InfoConstraint{
			Column: int(c.iColumn),
			Op:     Op(c.op),
			Usable: usable,
		})
	}
	return cst
}

func orderBys(info *C.sqlite3_index_info) []InfoOrderBy {
	slice := *(*[]C.struct_sqlite3_index_orderby)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(info.aOrderBy)),
		Len:  int(info.nOrderBy),
		Cap:  int(info.nOrderBy),
	}))

	ob := make([]InfoOrderBy, 0, len(slice))
	for _, c := range slice {
		var desc bool
		if c.desc > 0 {
			desc = true
		}
		ob = append(ob, InfoOrderBy{
			Column: int(c.iColumn),
			Desc:   desc,
		})
	}
	return ob
}

// IndexResult is a Go struct representation of what eventually ends up in the
// output fields for `sqlite3_index_info`
// See: https://www.sqlite.org/c3ref/index_info.html
type IndexResult struct {
	Used           []bool // aConstraintUsage
	IdxNum         int
	IdxStr         string
	AlreadyOrdered bool // orderByConsumed
	EstimatedCost  float64
	EstimatedRows  float64
}

// mPrintf is a utility wrapper around sqlite3_mprintf
func mPrintf(format, arg string) *C.char {
	cf := C.CString(format)
	defer C.free(unsafe.Pointer(cf))
	ca := C.CString(arg)
	defer C.free(unsafe.Pointer(ca))
	return C._sqlite3_mprintf(cf, ca)
}

//export goMInit
func goMInit(db, pClientData unsafe.Pointer, argc C.int, argv **C.char, pzErr **C.char, isCreate C.int) C.uintptr_t {
	m := lookupHandle(pClientData).(*sqliteModule)
	if m.c.db != (*C.sqlite3)(db) {
		*pzErr = mPrintf("%s", "Inconsistent db handles")
		return 0
	}
	args := make([]string, argc)
	var A []*C.char
	slice := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(argv)), Len: int(argc), Cap: int(argc)}
	a := reflect.NewAt(reflect.TypeOf(A), unsafe.Pointer(&slice)).Elem().Interface()
	for i, s := range a.([]*C.char) {
		args[i] = C.GoString(s)
	}
	var vTab VTab
	var err error
	if isCreate == 1 {
		vTab, err = m.module.Create(m.c, args)
	} else {
		vTab, err = m.module.Connect(m.c, args)
	}

	if err != nil {
		*pzErr = mPrintf("%s", err.Error())
		return 0
	}
	partialUpdate := false
	if up, ok := vTab.(VTabUpdater); ok {
		partialUpdate = up.PartialUpdate()
	}
	vt := sqliteVTab{m, vTab, partialUpdate}
	*pzErr = nil
	return C.uintptr_t(uintptr(newHandle(m.c, &vt)))
}

//export goVRelease
func goVRelease(pVTab unsafe.Pointer, isDestroy C.int) *C.char {
	vt := lookupHandle(pVTab).(*sqliteVTab)
	var err error
	if isDestroy == 1 {
		err = vt.vTab.Destroy()
	} else {
		err = vt.vTab.Disconnect()
	}
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVOpen
func goVOpen(pVTab unsafe.Pointer, pzErr **C.char) C.uintptr_t {
	vt := lookupHandle(pVTab).(*sqliteVTab)
	vTabCursor, err := vt.vTab.Open()
	if err != nil {
		*pzErr = mPrintf("%s", err.Error())
		return 0
	}
	vtc := sqliteVTabCursor{vt, vTabCursor, vt.partialUpdate}
	*pzErr = nil
	return C.uintptr_t(uintptr(newHandle(vt.module.c, &vtc)))
}

//export goVBestIndex
func goVBestIndex(pVTab unsafe.Pointer, icp unsafe.Pointer, pzErr **C.char) C.int {
	vt := lookupHandle(pVTab).(*sqliteVTab)
	info := (*C.sqlite3_index_info)(icp)
	csts := constraints(info)
	res, err := vt.vTab.BestIndex(csts, orderBys(info), IndexInformation{
		ColUsed: uint64(info.colUsed),
	})
	if err == ErrConstraint {
		return C.int(ErrConstraint)
	}

	if err != nil {
		if *pzErr != nil {
			C.sqlite3_free(unsafe.Pointer(*pzErr))
		}
		*pzErr = mPrintf("%s", err.Error())
		return C.SQLITE_ERROR
	}

	if len(res.Used) != len(csts) {
		return C.SQLITE_ERROR
	}

	// Get a pointer to constraint_usage struct so we can update in place.

	slice := *(*[]C.struct_sqlite3_index_constraint_usage)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(info.aConstraintUsage)),
		Len:  int(info.nConstraint),
		Cap:  int(info.nConstraint),
	}))
	index := 1
	for i := range slice {
		if res.Used[i] {
			// The default library omit value is 1
			// for used constraints
			// But anyquery uses constraints as a hint,
			// so it's still need SQLite to evaluate the constraints.
			// Therefore, we set omit to 0
			omit := C.uchar(0)

			// However, OFFSET is a special case because we can't offset the rows twice
			// Therefore, if an offset is used, we set omit to 1
			if csts[i].Op == OpOFFSET {
				omit = C.uchar(1)
			}

			slice[i].argvIndex = C.int(index)
			slice[i].omit = omit
			index++
		}
	}

	info.idxNum = C.int(res.IdxNum)
	info.idxStr = (*C.char)(C.sqlite3_malloc(C.int(len(res.IdxStr) + 1)))
	if info.idxStr == nil {
		// C.malloc and C.CString ordinarily do this for you. See https://golang.org/cmd/cgo/
		panic("out of memory")
	}
	info.needToFreeIdxStr = C.int(1)

	idxStr := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(info.idxStr)),
		Len:  len(res.IdxStr) + 1,
		Cap:  len(res.IdxStr) + 1,
	}))
	copy(idxStr, res.IdxStr)
	idxStr[len(idxStr)-1] = 0 // null-terminated string

	if res.AlreadyOrdered {
		info.orderByConsumed = C.int(1)
	}
	info.estimatedCost = C.double(res.EstimatedCost)
	info.estimatedRows = C.sqlite3_int64(res.EstimatedRows)

	return 0
}

//export goVClose
func goVClose(pCursor unsafe.Pointer) *C.char {
	vtc := lookupHandle(pCursor).(*sqliteVTabCursor)
	err := vtc.vTabCursor.Close()
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goMDestroy
func goMDestroy(pClientData unsafe.Pointer) {
	m := lookupHandle(pClientData).(*sqliteModule)
	m.module.DestroyModule()
}

//export goVFilter
func goVFilter(pCursor unsafe.Pointer, idxNum C.int, idxName *C.char, argc C.int, argv **C.sqlite3_value) *C.char {
	vtc := lookupHandle(pCursor).(*sqliteVTabCursor)
	args := (*[(math.MaxInt32 - 1) / unsafe.Sizeof((*C.sqlite3_value)(nil))]*C.sqlite3_value)(unsafe.Pointer(argv))[:argc:argc]
	vals := make([]any, 0, argc)
	for _, v := range args {
		conv, err := callbackArgGeneric(v)
		if err != nil {
			return mPrintf("%s", err.Error())
		}
		vals = append(vals, conv.Interface())
	}
	err := vtc.vTabCursor.Filter(int(idxNum), C.GoString(idxName), vals)
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVNext
func goVNext(pCursor unsafe.Pointer) *C.char {
	vtc := lookupHandle(pCursor).(*sqliteVTabCursor)
	err := vtc.vTabCursor.Next()
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVEof
func goVEof(pCursor unsafe.Pointer) C.int {
	vtc := lookupHandle(pCursor).(*sqliteVTabCursor)
	err := vtc.vTabCursor.EOF()
	if err {
		return 1
	}
	return 0
}

//export goVColumn
func goVColumn(pCursor, cp unsafe.Pointer, col C.int, nochange C.int) *C.char {
	vtc := lookupHandle(pCursor).(*sqliteVTabCursor)
	c := (*SQLiteContext)(cp)
	// When no change is set, it means we are in an update
	// and the column will not change.
	// Therefore, we can skip the column call
	if int(nochange) == 1 && vtc.partialUpdate {
		return nil
	}
	err := vtc.vTabCursor.Column(c, int(col))
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVRowid
func goVRowid(pCursor unsafe.Pointer, pRowid *C.sqlite3_int64) *C.char {
	vtc := lookupHandle(pCursor).(*sqliteVTabCursor)
	rowid, err := vtc.vTabCursor.Rowid()
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	*pRowid = C.sqlite3_int64(rowid)
	return nil
}

//export goVUpdate
func goVUpdate(pVTab unsafe.Pointer, argc C.int, argv **C.sqlite3_value, pRowid *C.sqlite3_int64) *C.char {
	vt := lookupHandle(pVTab).(*sqliteVTab)

	var tname string
	if n, ok := vt.vTab.(interface {
		TableName() string
	}); ok {
		tname = n.TableName() + " "
	}

	err := fmt.Errorf("virtual %s table %sis read-only", vt.module.name, tname)
	if v, ok := vt.vTab.(VTabUpdater); ok {
		// convert argv
		args := (*[(math.MaxInt32 - 1) / unsafe.Sizeof((*C.sqlite3_value)(nil))]*C.sqlite3_value)(unsafe.Pointer(argv))[:argc:argc]
		vals := make([]any, 0, argc)
		for _, v := range args {
			conv, err := callbackArgGeneric(v)
			if err != nil {
				return mPrintf("%s", err.Error())
			}

			// work around for SQLITE_NULL
			x := conv.Interface()
			if z, ok := x.([]byte); ok && z == nil {
				x = nil
			}

			vals = append(vals, x)
		}

		switch {
		case argc == 1:
			err = v.Delete(vals[0])

		case argc > 1 && vals[0] == nil:
			var id int64
			id, err = v.Insert(vals[1], vals[2:])
			if err == nil {
				*pRowid = C.sqlite3_int64(id)
			}

		case argc > 1:
			// We'll call sqlite3_value_nochange for each col
			// to ensure that the column is actually being updated
			// If not, we'll replace the value with nil
			// to indicate that the column should not be updated
			for i := 2; i < int(argc); i++ {
				if C.sqlite3_value_nochange(args[i]) == 1 {
					vals[i] = nil
				}
			}

			// In case of an update that changes the rowid, the rowid is passed as the first argument
			// If we need to access the new rowid, we can just find it in vals[2:]
			err = v.Update(vals[0], vals[2:])
		}
	} else {
		err = fmt.Errorf("virtual %s table %sis not updatable", vt.module.name, tname)
	}

	if err != nil {
		return mPrintf("%s", err.Error())
	}

	return nil
}

//export goVBegin
func goVBegin(pVTab unsafe.Pointer) *C.char {
	vt := lookupHandle(pVTab).(*sqliteVTab)
	if v, ok := vt.vTab.(VTabTransaction); ok {
		err := v.Begin()
		if err != nil {
			return mPrintf("%s", err.Error())
		}
	}
	return nil
}

//export goVCommit
func goVCommit(pVTab unsafe.Pointer) *C.char {
	vt := lookupHandle(pVTab).(*sqliteVTab)
	if v, ok := vt.vTab.(VTabTransaction); ok {
		err := v.Commit()
		if err != nil {
			return mPrintf("%s", err.Error())
		}
	}
	return nil
}

//export goVRollback
func goVRollback(pVTab unsafe.Pointer) *C.char {
	vt := lookupHandle(pVTab).(*sqliteVTab)
	if v, ok := vt.vTab.(VTabTransaction); ok {
		err := v.Rollback()
		if err != nil {
			return mPrintf("%s", err.Error())
		}
	}
	return nil
}

// Module is a "virtual table module", it defines the implementation of a
// virtual tables. See: http://sqlite.org/c3ref/module.html
type Module interface {
	// http://sqlite.org/vtab.html#xcreate
	Create(c *SQLiteConn, args []string) (VTab, error)
	// http://sqlite.org/vtab.html#xconnect
	Connect(c *SQLiteConn, args []string) (VTab, error)
	// http://sqlite.org/c3ref/create_module.html
	DestroyModule()
}

// EponymousOnlyModule is a "virtual table module" (as above), but
// for defining "eponymous only" virtual tables See: https://sqlite.org/vtab.html#eponymous_only_virtual_tables
type EponymousOnlyModule interface {
	Module
	EponymousOnlyModule()
}

// TransactionModule is a "virtual table module" (as above), but
// for defining virtual tables that support transactions.
type TransactionModule interface {
	Module
	TransactionModule()
}

type VTabTransaction interface {
	VTab
	Begin() error
	Commit() error
	Rollback() error
}

type IndexInformation struct {
	ColUsed uint64
}

// VTab describes a particular instance of the virtual table.
// See: http://sqlite.org/c3ref/vtab.html
type VTab interface {
	// http://sqlite.org/vtab.html#xbestindex
	BestIndex([]InfoConstraint, []InfoOrderBy, IndexInformation) (*IndexResult, error)
	// http://sqlite.org/vtab.html#xdisconnect
	Disconnect() error
	// http://sqlite.org/vtab.html#sqlite3_module.xDestroy
	Destroy() error
	// http://sqlite.org/vtab.html#xopen
	Open() (VTabCursor, error)
}

// VTabUpdater is a type that allows a VTab to be inserted, updated, or
// deleted.
// See: https://sqlite.org/vtab.html#xupdate
type VTabUpdater interface {
	VTab
	Delete(any) error
	Insert(any, []any) (int64, error)
	Update(any, []any) error
	PartialUpdate() bool
}

// VTabCursor describes cursors that point into the virtual table and are used
// to loop through the virtual table. See: http://sqlite.org/c3ref/vtab_cursor.html
type VTabCursor interface {
	// http://sqlite.org/vtab.html#xclose
	Close() error
	// http://sqlite.org/vtab.html#xfilter
	Filter(idxNum int, idxStr string, vals []any) error
	// http://sqlite.org/vtab.html#xnext
	Next() error
	// http://sqlite.org/vtab.html#xeof
	EOF() bool
	// http://sqlite.org/vtab.html#xcolumn
	Column(c *SQLiteContext, col int) error
	// http://sqlite.org/vtab.html#xrowid
	Rowid() (int64, error)
}

// DeclareVTab declares the Schema of a virtual table.
// See: http://sqlite.org/c3ref/declare_vtab.html
func (c *SQLiteConn) DeclareVTab(sql string) error {
	zSQL := C.CString(sql)
	defer C.free(unsafe.Pointer(zSQL))
	rv := C.sqlite3_declare_vtab(c.db, zSQL)
	if rv != C.SQLITE_OK {
		return c.lastError()
	}
	return nil
}

// CreateModule registers a virtual table implementation.
// See: http://sqlite.org/c3ref/create_module.html
func (c *SQLiteConn) CreateModule(moduleName string, module Module) error {
	mname := C.CString(moduleName)
	defer C.free(unsafe.Pointer(mname))
	udm := sqliteModule{c, moduleName, module}
	switch module.(type) {
	case TransactionModule:
		rv := C._sqlite3_create_module_transaction(c.db, mname, C.uintptr_t(uintptr(newHandle(c, &udm))))
		if rv != C.SQLITE_OK {
			return c.lastError()
		}
		return nil

	case EponymousOnlyModule:
		rv := C._sqlite3_create_module_eponymous_only(c.db, mname, C.uintptr_t(uintptr(newHandle(c, &udm))))
		if rv != C.SQLITE_OK {
			return c.lastError()
		}
		return nil
	case Module:
		rv := C._sqlite3_create_module(c.db, mname, C.uintptr_t(uintptr(newHandle(c, &udm))))
		if rv != C.SQLITE_OK {
			return c.lastError()
		}
		return nil
	}
	return nil
}

// Drop a module by its name
func (c *SQLiteConn) DropModule(moduleName string) error {
	// SQLite doesn't provide a way to drop a module by name.
	// However, SQLite has a method sqlite3_drop_modules
	// that drops all the modules except the ones in a specified list.
	//
	// This function uses this feature. It queries all the modules,
	// add them to the skip list, and then calls sqlite3_drop_modules.

	// Get all the module names
	var keep []*C.char
	rows, err := c.Query("PRAGMA module_list;", []driver.Value{})
	if err != nil {
		return err
	}

	moduleNameQuery := make([]driver.Value, 1)
	for {
		clear(moduleNameQuery)
		err = rows.Next(moduleNameQuery)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(moduleNameQuery) != 1 {
			return fmt.Errorf("unexpected number of columns in pragma_module_list (expected 1, got %d)", len(moduleNameQuery))
		}

		name, ok := moduleNameQuery[0].(string)
		if !ok {
			return fmt.Errorf("unexpected type of column in pragma_module_list (expected string, got %T)", moduleNameQuery[0])
		}

		// Add all module except the one we want to drop
		if name != moduleName {
			keep = append(keep, C.CString(name))
		}
	}
	rows.Close()

	// As specified in the documentation, the last element of the array must be NULL
	keep = append(keep, nil)

	// Drop all the modules except the ones in the keep list
	rv := C._sqlite3_drop_modules(c.db, &keep[0])

	if rv != C.SQLITE_OK {
		return c.lastError()
	}

	return nil
}
