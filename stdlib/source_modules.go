// gensrcmods.go ile oluşturuldu, değiştirmeyin.

package stdlib

var SourceModules = map[string]string{
	"enum": "is_enumerable := fn(x) {\r\n  dön is_array(x) || is_map(x) || is_immutable_array(x) || is_immutable_map(x)\r\n}\r\n\r\nis_array_like := fn(x) {\r\n  dön is_array(x) || is_immutable_array(x)\r\n}\r\n\r\npaylaş {\r\n  all: fn(x, fn) {\r\n    if !is_enumerable(x) { dön undefined }\r\n    for k, v in x {\r\n      if !fn(k, v) { dön yanlış }\r\n    }\r\n    dön doğru\r\n  },\r\n  any: fn(x, fn) {\r\n    if !is_enumerable(x) { dön undefined }\r\n    for k, v in x {\r\n      if fn(k, v) { dön doğru }\r\n    }\r\n    dön yanlış\r\n  },\r\n\r\n  chunk: fn(x, size) {\r\n    if !is_array_like(x) || !size { dön undefined }\r\n    numElements := len(x)\r\n    if !numElements { dön [] }\r\n    res := []\r\n    idx := 0\r\n    for idx < numElements {\r\n      res = append(res, x[idx:idx+size])\r\n      idx += size\r\n    }\r\n    dön res\r\n  },\r\n  at: fn(x, key) {\r\n    if !is_enumerable(x) { dön undefined }\r\n    if is_array_like(x) {\r\n        if !is_int(key) { dön undefined }\r\n    } else {\r\n        if !is_string(key) { dön undefined }\r\n    }\r\n    dön x[key]\r\n  },\r\n  \r\n  each: fn(x, fn) {\r\n    if !is_enumerable(x) { dön undefined }\r\n    for k, v in x {\r\n      fn(k, v)\r\n    }\r\n  },\r\n  \r\n  filter: fn(x, fn) {\r\n    if !is_array_like(x) { dön undefined }\r\n    dst := []\r\n    for k, v in x {\r\n      if fn(k, v) { dst = append(dst, v) }\r\n    }\r\n    dön dst\r\n  },\r\n  \r\n  find: fn(x, fn) {\r\n    if !is_enumerable(x) { dön undefined }\r\n    for k, v in x {\r\n      if fn(k, v) { dön v }\r\n    }\r\n  },\r\n  \r\n  find_key: fn(x, fn) {\r\n    if !is_enumerable(x) { dön undefined }\r\n    for k, v in x {\r\n      if fn(k, v) { dön k }\r\n    }\r\n  },\r\n  \r\n  map: fn(x, fn) {\r\n    if !is_enumerable(x) { dön undefined }\r\n    dst := []\r\n    for k, v in x {\r\n      dst = append(dst, fn(k, v))\r\n    }\r\n    dön dst\r\n  },\r\n  \r\n  key: fn(k, _) { dön k },\r\n  \r\n  value: fn(_, v) { dön v }\r\n}",
}