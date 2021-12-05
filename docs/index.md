


# Action Runner Manager API
API for managing GitHub organization Runner Groups
  

## Informations

### Version

0.1.0

### License

[Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0.html)

### Contact

GitHub Professional Services lindluni@github.com https://github.com/lindluni/actions-runner-manager

## Content negotiation

### URI Schemes
  * http

### Consumes
  * application/json

### Produces
  * application/json

## All endpoints

###  groups

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/groups-create/{team} | [get groups create team](#get-groups-create-team) | Create a new GitHub Action organization Runner Group |
| GET | /api/v1/groups-delete/{team} | [get groups delete team](#get-groups-delete-team) | Deletes an existing GitHub Action organization Runner Group |
| GET | /api/v1/groups-list/{team}:{repos} | [get groups list team repos](#get-groups-list-team-repos) | List all resources associated with a GitHub Action organization Runner Group |
  


###  tokens

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/token-register/{team} | [get token register team](#get-token-register-team) | Create a new GitHub Action organization runner registration token |
| GET | /api/v1/token-remove/{team} | [get token remove team](#get-token-remove-team) | Create a new GitHub Action organization runner removal token |
  


## Paths

### <span id="get-groups-create-team"></span> Create a new GitHub Action organization Runner Group (*GetGroupsCreateTeam*)

```
GET /api/v1/groups-create/{team}
```

Creates a new GitHub Action organization runner group named with the team slug

#### Produces
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| team | `path` | string | `string` |  | ✓ |  | Canonical **slug** of the GitHub team |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-groups-create-team-200) | OK | OK |  | [schema](#get-groups-create-team-200-schema) |

#### Responses


##### <span id="get-groups-create-team-200"></span> 200 - OK
Status: OK

###### <span id="get-groups-create-team-200-schema"></span> Schema
   
  

[GetGroupsCreateTeamOKBody](#get-groups-create-team-o-k-body)

###### Inlined models

**<span id="get-groups-create-team-o-k-body"></span> GetGroupsCreateTeamOKBody**


  


* composed type [ApisJSONResult](#apis-json-result)
* inlined member (*getGroupsCreateTeamOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| Code | integer| `int64` |  | |  |  |
| Response | string| `string` |  | |  |  |



### <span id="get-groups-delete-team"></span> Deletes an existing GitHub Action organization Runner Group (*GetGroupsDeleteTeam*)

```
GET /api/v1/groups-delete/{team}
```

Deletes an existing GitHub Action organization runner group named with the team slug

#### Produces
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| team | `path` | string | `string` |  | ✓ |  | Canonical **slug** of the GitHub team |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-groups-delete-team-200) | OK | OK |  | [schema](#get-groups-delete-team-200-schema) |

#### Responses


##### <span id="get-groups-delete-team-200"></span> 200 - OK
Status: OK

###### <span id="get-groups-delete-team-200-schema"></span> Schema
   
  

[GetGroupsDeleteTeamOKBody](#get-groups-delete-team-o-k-body)

###### Inlined models

**<span id="get-groups-delete-team-o-k-body"></span> GetGroupsDeleteTeamOKBody**


  


* composed type [ApisJSONResult](#apis-json-result)
* inlined member (*getGroupsDeleteTeamOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| Code | integer| `int64` |  | |  |  |
| Response | string| `string` |  | |  |  |



### <span id="get-groups-list-team-repos"></span> List all resources associated with a GitHub Action organization Runner Group (*GetGroupsListTeamRepos*)

```
GET /api/v1/groups-list/{team}:{repos}
```

List all repositories and runners assigned to a GitHub Action organization runner group named with the team slug

#### Produces
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| repos | `path` | []string | `[]string` |  | ✓ |  | Comma-seperated list of repository slugs |
| team | `path` | string | `string` |  | ✓ |  | Canonical **slug** of the GitHub team |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-groups-list-team-repos-200) | OK | OK |  | [schema](#get-groups-list-team-repos-200-schema) |

#### Responses


##### <span id="get-groups-list-team-repos-200"></span> 200 - OK
Status: OK

###### <span id="get-groups-list-team-repos-200-schema"></span> Schema
   
  

[GetGroupsListTeamReposOKBody](#get-groups-list-team-repos-o-k-body)

###### Inlined models

**<span id="get-groups-list-team-repos-o-k-body"></span> GetGroupsListTeamReposOKBody**


  


* composed type [ApisJSONResult](#apis-json-result)
* inlined member (*getGroupsListTeamReposOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| Code | integer| `int64` |  | |  |  |
| Response | [ApisListResponse](#apis-list-response)| `models.ApisListResponse` |  | |  |  |



### <span id="get-token-register-team"></span> Create a new GitHub Action organization runner registration token (*GetTokenRegisterTeam*)

```
GET /api/v1/token-register/{team}
```

Creates a new GitHub Action organization runner removal token that can be used to configure GitHub Action runners at the organization level

#### Produces
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| team | `path` | string | `string` |  | ✓ |  | Canonical **slug** of the GitHub team |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-token-register-team-200) | OK | OK |  | [schema](#get-token-register-team-200-schema) |

#### Responses


##### <span id="get-token-register-team-200"></span> 200 - OK
Status: OK

###### <span id="get-token-register-team-200-schema"></span> Schema
   
  

[GetTokenRegisterTeamOKBody](#get-token-register-team-o-k-body)

###### Inlined models

**<span id="get-token-register-team-o-k-body"></span> GetTokenRegisterTeamOKBody**


  


* composed type [ApisJSONResult](#apis-json-result)
* inlined member (*getTokenRegisterTeamOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| Code | integer| `int64` |  | |  |  |
| Response | [GithubRegistrationToken](#github-registration-token)| `models.GithubRegistrationToken` |  | |  |  |



### <span id="get-token-remove-team"></span> Create a new GitHub Action organization runner removal token (*GetTokenRemoveTeam*)

```
GET /api/v1/token-remove/{team}
```

Creates a new GitHub Action organization runner removal token that can be used remove a GitHub Action runners at the organization level

#### Produces
  * application/json

#### Security Requirements
  * ApiKeyAuth

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| team | `path` | string | `string` |  | ✓ |  | Canonical **slug** of the GitHub team |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-token-remove-team-200) | OK | OK |  | [schema](#get-token-remove-team-200-schema) |

#### Responses


##### <span id="get-token-remove-team-200"></span> 200 - OK
Status: OK

###### <span id="get-token-remove-team-200-schema"></span> Schema
   
  

[GetTokenRemoveTeamOKBody](#get-token-remove-team-o-k-body)

###### Inlined models

**<span id="get-token-remove-team-o-k-body"></span> GetTokenRemoveTeamOKBody**


  


* composed type [ApisJSONResult](#apis-json-result)
* inlined member (*getTokenRemoveTeamOKBodyAO1*)



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| Code | integer| `int64` |  | |  |  |
| Response | [GithubRegistrationToken](#github-registration-token)| `models.GithubRegistrationToken` |  | |  |  |



## Models

### <span id="apis-json-result"></span> apis.JSONResult


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| Code | integer| `int64` |  | |  |  |
| Message | string| `string` |  | |  |  |
| Response | [interface{}](#interface)| `interface{}` |  | |  |  |



### <span id="apis-list-response"></span> apis.listResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| repos | []string| `[]string` |  | |  |  |
| runners | []string| `[]string` |  | |  |  |



### <span id="github-registration-token"></span> github.RegistrationToken


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| expires_at | [GithubTimestamp](#github-timestamp)| `GithubTimestamp` |  | |  |  |
| token | string| `string` |  | |  |  |



### <span id="github-timestamp"></span> github.Timestamp


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| time.Time | string| `string` |  | |  |  |


