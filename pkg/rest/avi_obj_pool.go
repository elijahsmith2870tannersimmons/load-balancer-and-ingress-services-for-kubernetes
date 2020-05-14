/*
 * [2013] - [2019] Avi Networks Incorporated
 * All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *   http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rest

import (
	"encoding/json"
	"errors"
	"fmt"

	avicache "ako/pkg/cache"
	"ako/pkg/nodes"

	"github.com/avinetworks/container-lib/utils"
	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

func (rest *RestOperations) AviPoolBuild(pool_meta *nodes.AviPoolNode, cache_obj *avicache.AviPoolCache, key string) *utils.RestOp {
	name := pool_meta.Name
	cksum := pool_meta.CloudConfigCksum
	cksumString := fmt.Sprint(cksum)
	tenant := fmt.Sprintf("/api/tenant/?name=%s", pool_meta.Tenant)
	cr := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	svc_mdata_json, _ := json.Marshal(&pool_meta.ServiceMetadata)
	svc_mdata := string(svc_mdata_json)
	cloudRef := "/api/cloud?name=" + utils.CloudName
	vrfContextRef := "/api/vrfcontext?name=" + pool_meta.VrfContext
	pool := avimodels.Pool{Name: &name, CloudConfigCksum: &cksumString,
		CreatedBy: &cr, TenantRef: &tenant, CloudRef: &cloudRef,
		ServiceMetadata: &svc_mdata, VrfRef: &vrfContextRef}

	for _, server := range pool_meta.Servers {
		sip := server.Ip
		port := pool_meta.Port
		s := avimodels.Server{IP: &sip, Port: &port}
		if server.ServerNode != "" {
			sn := server.ServerNode
			s.ServerNode = &sn
		}
		pool.Servers = append(pool.Servers, &s)
	}

	var hm string
	if pool_meta.Protocol == "udp" {
		hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_UDP_HM)
	} else {
		hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_TCP_HM)
	}
	pool.HealthMonitorRefs = append(pool.HealthMonitorRefs, hm)

	macro := utils.AviRestObjMacro{ModelName: "Pool", Data: pool}

	// TODO Version should be latest from configmap
	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/pool/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pool,
			Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}
	} else {
		// Patch an existing pool if it exists in the cache but not associated with this VS.
		pool_key := avicache.NamespaceName{Namespace: pool_meta.Tenant, Name: name}
		pool_cache, ok := rest.cache.PoolCache.AviCacheGet(pool_key)
		if ok {
			pool_cache_obj, _ := pool_cache.(*avicache.AviPoolCache)
			path = "/api/pool/" + pool_cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pool,
				Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}
		} else {
			path = "/api/macro"
			rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
				Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}
		}
	}

	utils.AviLog.Debug(spew.Sprintf("key: %s, msg: pool Restop %v K8sAviPoolMeta %v\n", key,
		utils.Stringify(rest_op), *pool_meta))
	return &rest_op
}

func (rest *RestOperations) AviPoolDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/pool/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "Pool", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("key: %s, msg: pool DELETE Restop %v \n", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviPoolCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no reponse for POOL, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "pool", key)
	utils.AviLog.Debugf("key: %s, msg: the pool object response %v", key, rest_op.Response)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find pool obj in resp %v", key, rest_op.Response)
		return errors.New("pool not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: uuid not present in response %v", key, resp)
			continue
		}
		cksum := resp["cloud_config_cksum"].(string)

		var svc_mdata_obj avicache.ServiceMetadataObj
		if resp["service_metadata"] != nil {
			if err := json.Unmarshal([]byte(resp["service_metadata"].(string)),
				&svc_mdata_obj); err != nil {
				utils.AviLog.Warnf("Error parsing service metadata :%v", err)
			}
		}

		var lastModifiedStr string
		lastModifiedIntf, ok := resp["_last_modified"]
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: last_modified not present in response %v", key, resp)
		} else {
			lastModifiedStr, ok = lastModifiedIntf.(string)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: last_modified is not of type string", key)
			}
		}

		pool_cache_obj := avicache.AviPoolCache{Name: name, Tenant: rest_op.Tenant,
			Uuid:               uuid,
			CloudConfigCksum:   cksum,
			ServiceMetadataObj: svc_mdata_obj,
			LastModified:       lastModifiedStr,
		}
		if lastModifiedStr == "" {
			pool_cache_obj.InvalidData = true
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.PoolCache.AviCacheAdd(k, &pool_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCache.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToPoolKeyCollection(k)
				utils.AviLog.Debugf("key: %s, msg: modified the VS cache object for Pool Collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
				if svc_mdata_obj.Namespace != "" {
					UpdateIngressStatus(vs_cache_obj, svc_mdata_obj, key)
				}
			}
		} else {
			vs_cache_obj := rest.cache.VsCache.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToPoolKeyCollection(k)
			utils.AviLog.Debug(spew.Sprintf("key: %s, msg: added VS cache key during pool update %v val %v\n", key, vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Info(spew.Sprintf("key: %s, msg: Added Pool cache k %v val %v\n", key, k,
			pool_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviPoolCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	poolKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	utils.AviLog.Debug("key: %s, msg: deleting pool with key: %s", key, poolKey)
	// Fetch the pool's cache data and obtain the service metadata
	pool_cache, found := rest.cache.PoolCache.AviCacheGet(poolKey)
	if found {
		pool_cache_obj, success := pool_cache.(*avicache.AviPoolCache)
		if success {
			err := DeleteIngressStatus(pool_cache_obj.ServiceMetadataObj, key)
			if k8serror.IsNotFound(err) {
				// Just log and get away
				utils.AviLog.Infof("key: %s, msg: ingress already deleted, nothing to update in status", key)
			}
		}
	}
	// Now delete the cache.
	rest.cache.PoolCache.AviCacheDelete(poolKey)
	// Delete the pool from the vs cache as well.
	vs_cache, ok := rest.cache.VsCache.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			utils.AviLog.Debugf("key: %s, msg: VsKey: %s, VS Pool key cache before deletion :%s", key, vsKey, vs_cache_obj.PoolKeyCollection)
			vs_cache_obj.RemoveFromPoolKeyCollection(poolKey)
			utils.AviLog.Infof("key: %s, msg: VS Pool key cache after deletion :%s", key, vs_cache_obj.PoolKeyCollection)
		}
	}

	return nil
}
