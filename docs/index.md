


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

###  tokens

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/token-register/{team} | [get token register team](#get-token-register-team) | Creates a new GitHub Action organization runner registration token |
| GET | /api/v1/token-remove/{team} | [get token remove team](#get-token-remove-team) | Creates a new GitHub Action organization runner removal token |
  


## Paths

### <span id="get-token-register-team"></span> Creates a new GitHub Action organization runner registration token (*GetTokenRegisterTeam*)

```
GET /api/v1/token-register/{team}
```

Create a new GitHub Action organization runner registration token

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
   
  

[GithubRegistrationToken](#github-registration-token)

### <span id="get-token-remove-team"></span> Creates a new GitHub Action organization runner removal token (*GetTokenRemoveTeam*)

```
GET /api/v1/token-remove/{team}
```

Create a new GitHub Action organization runner removal token

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
   
  

[GithubRegistrationToken](#github-registration-token)

## Models

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


