# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [article_service.proto](#article_service-proto)
    - [Article](#article-Article)
    - [CreateArticleRequest](#article-CreateArticleRequest)
    - [CreateArticleResponse](#article-CreateArticleResponse)
    - [DeleteArticleRequest](#article-DeleteArticleRequest)
    - [GetArticleRequest](#article-GetArticleRequest)
    - [GetArticleResponse](#article-GetArticleResponse)
    - [GetArticlesRequest](#article-GetArticlesRequest)
    - [UpdateArticleRequest](#article-UpdateArticleRequest)
  
    - [ArticleService](#article-ArticleService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="article_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## article_service.proto



<a name="article-Article"></a>

### Article



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| title | [string](#string) |  |  |
| body | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="article-CreateArticleRequest"></a>

### CreateArticleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| article | [Article](#article-Article) |  |  |






<a name="article-CreateArticleResponse"></a>

### CreateArticleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="article-DeleteArticleRequest"></a>

### DeleteArticleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="article-GetArticleRequest"></a>

### GetArticleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="article-GetArticleResponse"></a>

### GetArticleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| article | [Article](#article-Article) |  |  |






<a name="article-GetArticlesRequest"></a>

### GetArticlesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| offset | [string](#string) |  |  |
| limit | [string](#string) |  |  |






<a name="article-UpdateArticleRequest"></a>

### UpdateArticleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| article | [Article](#article-Article) |  |  |





 

 

 


<a name="article-ArticleService"></a>

### ArticleService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Create | [CreateArticleRequest](#article-CreateArticleRequest) | [CreateArticleResponse](#article-CreateArticleResponse) |  |
| Update | [UpdateArticleRequest](#article-UpdateArticleRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| Delete | [DeleteArticleRequest](#article-DeleteArticleRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| Get | [GetArticleRequest](#article-GetArticleRequest) | [GetArticleResponse](#article-GetArticleResponse) |  |
| GetStream | [GetArticlesRequest](#article-GetArticlesRequest) | [Article](#article-Article) stream |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

