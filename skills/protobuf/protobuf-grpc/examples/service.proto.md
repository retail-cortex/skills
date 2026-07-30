# Package: enterprise.customer.v1

<div class="comment"><span></span><br/></div>

## Imports

| Import | Description |
|--------|-------------|



## Options

| Name                | Value                                         | Description |
|---------------------|-----------------------------------------------|-------------|
| go_package          | github.com/enterprise/service/api/customer/v1 |             |
| java_multiple_files | true                                          |             |
| java_package        | com.enterprise.customer.v1                    |             |



### enterprise.customer.v1 Diagram

```mermaid
classDiagram
direction LR
%% Mermaid Diagram for package: enterprise.customer.v1

%% 

class GetCustomerRequest {
  + string customer_id
}

%% 

class ListCustomersRequest {
  + int32 page_size
  + string page_token
}

%% 

class CustomerResponse {
  + string customer_id
  + string name
  + string email
  + bool is_active
}

%% 

class ListCustomersResponse {
  + List~CustomerResponse~ customers
  + string next_page_token
}
ListCustomersResponse --> `CustomerResponse`
class CustomerService {
  <<service>>
  +GetCustomer (GetCustomerRequest) CustomerResponse
  +ListCustomers (ListCustomersRequest) ListCustomersResponse
}
CustomerService --> `GetCustomerRequest`
CustomerService --> `CustomerResponse`
CustomerService --> `ListCustomersRequest`
CustomerService --> `ListCustomersResponse`

```

## Service: CustomerService
<div style="font-size: 12px; margin-top: -10px;" class="fqn">FQN: enterprise.customer.v1</div>

<div class="comment"><span></span><br/></div>

### CustomerService Diagram

```mermaid
classDiagram
direction LR
class CustomerService {
  <<service>>
  +GetCustomer (GetCustomerRequest) CustomerResponse
  +ListCustomers (ListCustomersRequest) ListCustomersResponse
}
CustomerService --> `GetCustomerRequest`
CustomerService --> `CustomerResponse`
CustomerService --> `ListCustomersRequest`
CustomerService --> `ListCustomersResponse`

```

| Method         | Parameter (In)       | Parameter (Out)       | Description |
|----------------|----------------------|-----------------------|-------------|
| GetCustomer    | GetCustomerRequest   | CustomerResponse      |             |
| ListCustomers  | ListCustomersRequest | ListCustomersResponse |             |



### GetCustomerRequest Diagram

```mermaid
classDiagram
direction LR

%% 

class GetCustomerRequest {
  + string customer_id
}

```
### ListCustomersRequest Diagram

```mermaid
classDiagram
direction LR

%% 

class ListCustomersRequest {
  + int32 page_size
  + string page_token
}

```
### CustomerResponse Diagram

```mermaid
classDiagram
direction LR

%% 

class CustomerResponse {
  + string customer_id
  + string name
  + string email
  + bool is_active
}

```
### ListCustomersResponse Diagram

```mermaid
classDiagram
direction LR

%% 

class ListCustomersResponse {
  + List~CustomerResponse~ customers
  + string next_page_token
}
ListCustomersResponse --> `CustomerResponse`

```

## Message: GetCustomerRequest
<div style="font-size: 12px; margin-top: -10px;" class="fqn">FQN: enterprise.customer.v1.GetCustomerRequest</div>

<div class="comment"><span></span><br/></div>

| Field       | Ordinal | Type   | Label | Description |
|-------------|---------|--------|-------|-------------|
| customer_id | 1       | string |       |             |




## Message: ListCustomersRequest
<div style="font-size: 12px; margin-top: -10px;" class="fqn">FQN: enterprise.customer.v1.ListCustomersRequest</div>

<div class="comment"><span></span><br/></div>

| Field      | Ordinal | Type   | Label | Description |
|------------|---------|--------|-------|-------------|
| page_size  | 1       | int32  |       |             |
| page_token | 2       | string |       |             |




## Message: CustomerResponse
<div style="font-size: 12px; margin-top: -10px;" class="fqn">FQN: enterprise.customer.v1.CustomerResponse</div>

<div class="comment"><span></span><br/></div>

| Field       | Ordinal | Type   | Label | Description |
|-------------|---------|--------|-------|-------------|
| customer_id | 1       | string |       |             |
| name        | 2       | string |       |             |
| email       | 3       | string |       |             |
| is_active   | 4       | bool   |       |             |




## Message: ListCustomersResponse
<div style="font-size: 12px; margin-top: -10px;" class="fqn">FQN: enterprise.customer.v1.ListCustomersResponse</div>

<div class="comment"><span></span><br/></div>

| Field           | Ordinal | Type             | Label    | Description |
|-----------------|---------|------------------|----------|-------------|
| customers       | 1       | CustomerResponse | Repeated |             |
| next_page_token | 2       | string           |          |             |






<!-- Created by: Proto Diagram Tool -->
<!-- https://github.com/GoogleCloudPlatform/proto-gen-md-diagrams -->
