/*
 * [2013] - [2018] Avi Networks Incorporated
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

package utils

const (
	GraphLayer                    = "GraphLayer"
	ObjectIngestionLayer          = "ObjectIngestionLayer"
	LeastConnection               = "LB_ALGORITHM_LEAST_CONNECTIONS"
	RandomConnection              = "RANDOM_CONN"
	PassthroughConnection         = "PASSTHROUGH_CONN"
	RoundRobinConnection          = "LB_ALGORITHM_ROUND_ROBIN"
	ServiceInformer               = "ServiceInformer"
	PodInformer                   = "PodInformer"
	SecretInformer                = "SecretInformer"
	EndpointInformer              = "EndpointInformer"
	K8S_TLS_SECRET_CERT           = "tls.cert"
	K8S_TLS_SECRET_KEY            = "tls.key"
	IngressInformer               = "IngressInformer"
	RouteInformer                 = "RouteInformer"
	L4LBService                   = "L4LBService"
	LoadBalancer                  = "LoadBalancer"
	Endpoints                     = "Endpoints"
	Ingress                       = "Ingress"
	Service                       = "Service"
	HTTP                          = "HTTP"
	HeaderMethod                  = ":method"
	HeaderAuthority               = ":authority"
	HeaderScheme                  = ":scheme"
	TLS                           = "TLS"
	HTTPS                         = "HTTPS"
	TCP                           = "TCP"
	UDP                           = "UDP"
	SYSTEM_UDP_FAST_PATH          = "System-UDP-Fast-Path"
	DEFAULT_TCP_NW_PROFILE        = "System-TCP-Proxy"
	DEFAULT_L4_APP_PROFILE        = "System-L4-Application"
	DEFAULT_L7_APP_PROFILE        = "System-HTTP"
	DEFAULT_L7_SECURE_APP_PROFILE = "System-Secure-HTTP"
	DEFAULT_SHARD_VS_PREFIX       = "Shard-VS-"
	L7_PG_PREFIX                  = "-PG-l7"
	VS_DATASCRIPT_EVT_HTTP_REQ    = "VS_DATASCRIPT_EVT_HTTP_REQ"
	HTTP_DS_SCRIPT                = "host = avi.http.get_host_tokens(1)\npath = avi.http.get_path_tokens(1)\nif host and path then\nlbl = host..\"/\"..path\nelse\nlbl = host..\"/\"\nend\navi.poolgroup.select(\"POOLGROUP\", string.lower(lbl) )"
	ADMIN_NS                      = "admin"
	TLS_PASSTHROUGH               = "TLS_PASSTHROUGH"
	VS_TYPE_VH_PARENT             = "VS_TYPE_VH_PARENT"
	VS_TYPE_VH_CHILD              = "VS_TYPE_VH_CHILD"
)
