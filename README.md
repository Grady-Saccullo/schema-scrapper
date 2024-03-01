# Name TBD bc don't care 🤌

Parsing html sucks. This is a tool to make it suck less, lets be clear it still sucks tho...

This tool takes in html and an output schema and returns the correctly built json.


> This is very much so a work in progress, however the end goal is to have a library which can take in html and
> a given schema and return the correctly built json.

## Motivation

A friend/colleague from [Seer](https://withseer.ai/) reached out to me and asked if I knew of a library which could
parse html into json, but via a schema definition/file so they could control the output shape.
I didn't, so I said hold my 🍺 and here we are.

While building this it did occur how useful this could be for apis/applications which need to scrape data from
websites. As someone who has had to write website scrappers before, I could see how using a schema file to define
the transformation from html to json could be a lot more maintainable. So long as the output object is close
enough to what you need, writing small code to do the last transform would be easy.

### Seer Use Case
They need to get legislative session data from state government websites. This data is not available via an API the
vast majority of the time, or if it it it's not reliable/up to date. So they need to scrape the data from the website.
This is a huge pain in the 🍑 because each state has a different website and each website has a different structure
as well as each session might get tweaked slightly with html structure. This means they need to write a custom scraper
for each state and potentially each session. Being able to define a schema for each state and session allows them to
only need to maintain a schema and not custom code. This is a huge time saver and allows them to scale much more
efficiently.

---


## The Node Object

The "node" object is the core of the schema. Nodes can be used in a variety of ways to select and parse data from the html.


### Selector `$selector`

The selector is the most important part of the node. It is used to select the html node from which to parse the data.
In the examples below it should look very similar to a css selector, as this is what it is based off of.

Selectors are made up of a few parts:

#### Tag Name `body.div.p`

The tag name (div, p, a, etc) in dot notation from the root of the html.

```json
{
    "$type": "string",
    "$selector": "body.div.p"
}
```

#### Attributes `[ ... ]`

This is a list of key value pairs separated by commas. The key is the attribute name and the value is the value of the attribute. This is used to filter the tags selected by the tag name.

**Match against value**
- `class=subtext`
- `id=unique`
- `some-custom-attribute=some-value`

**Check for key existence**
- `[x-is-first=true]`
- `[x-is-last=false]`


There are some reserved attributes which have special meaning. All reserved attributes start with `!` to help
differentiate them from custom attributes and prevent any conflicts.

- `!nth` - The index of the tag to select. This is 0 based or first/last. Currently only accepts a number.

```json 
{
    "class_select": {
        "$type": "string",
        "$selector": "div.p[class=subtext]"
    },
    "id_select": {
        "$type": "string",
        "$selector": "div.p[id=unique]"
    },
    "non_standard_attribute": {
        "$type": "string",
        "$selector": "div.p[x-data-test-id=123]"
    }
    "nth_select": {
        "$type": "string",
        "$selector": "div.p[!nth=2]"
    },
    "compound_select": {
        "$type": "string",
        "$selector": "div.p[!nth=2,class=subtext]"
    }
}
```

#### Get Data from Attributes `:<attribute-name>` (WIP/Not Implemented)

This is a way to get the value of an attribute from the selected tag to use as the value of the node.

This looks very similar to psuedo selectors in css. For the current version of this library no psuedo
selectors are supported which is why we are abusing the syntax.

```json
{
    "href": {
        "$type": "string",
        "$selector": "div.a:href"
    }
}
```



### Supported JSON Types:

| Type      | Supported |
|-----------|-----------|
| `string`  | ✅        |
| `number`  | ✅        |
| `boolean` | ❌ (soon) |
| `object`  | ✅        |
| `array`   | ✅        |
| `null`    | ❌ (soon) |

### `string`

| Field       | Required | Description |
|-------------|----------|-------------|
| `$type`     | ✅       | `string`    |
| `$selector` | ✅       | The node to select from the html. |

**Input HTML**
```html
<html>
    <div>
        <p>Turing</p>
        <p class="subtext">Alan Turing</h2>
        <div class="age">
            <p>age</p>
            <p>41</p>
            <p>years</p>
        </div>
        <div>
            <a href="https://en.wikipedia.org/wiki/Alan_Turing">Wikipedia</a>
        </div>
    </div>
</html>
```

**Input Schema**
```json
{
    "name": {
        "$type": "string",
        "$selector": "div.p[nth=0]",
    },
    "subtext": {
        "$type": "string",
        "$selector": "div.p[class=subtext]"
    },
    "link": {
        "$type": "string",
        "$selector": "div[nth=0].a:href"
    },
    "data": {
        "name": {
            "$type": "string",
            "$selector": "div.p[nth=0]",
        },
        "subtext": {
            "$type": "string",
            "$selector": "div.p[class=subtext]"
        },
        "link": {
            "$type": "string",
            "$selector": "div[nth=1].a:href"
        }
    }
}
```

**Output JSON**
```json
{
    "name": "Turing",
    "subtext": "Alan Turing",
    "link": "https://en.wikipedia.org/wiki/Alan_Turing",
    "data": {
        "name": "Turing",
        "subtext": "Alan Turing",
        "link": "https://en.wikipedia.org/wiki/Alan_Turing"
    }
}
```

### `number`

This will output the value of the selected node as a number


| Field       | Required | Description |
|-------------|----------|-------------|
| `$type`     | ✅       | `number`    |
| `$selector` | ✅       | The node to select from the html. |

**Input HTML**
```html
<html>
    <div>
        <p>Turing</p>
        <p class="subtext">Alan Turing</h2>
        <div class="age">
            <p>age</p>
            <p>41</p>
            <p>years</p>
        </div>
        <div>
            <a href="https://en.wikipedia.org/wiki/Alan_Turing">Wikipedia</a>
        </div>
    </div>
</html>
```

**Input Schema**
```json
{
    "age": {
        "$type": "number",
        "$selector": "div",
        "$index": 1
    }
}
```

**Output JSON**
```json
{
    "age": 41
}
```


### `array`

| Field           | Required | Description |
|-----------------|----------|-------------|
| `$type`         | ✅       | `array` |
| `$value_node`   | ✅       | The node to select for the object value. Can be any type of node. |
| `$itr_el`       | ✅       | The node to iterate over to build the array. |
| `$itr_ident`    | ✅       | The internal identifier to reference in the schema. |
| `$itr_el_match` | ❌       | The tag to match against when iterating. This has the same power as the `$selector` field. If not supplied assumes `*` |
| `$start_offset` | ❌       | The index to start the iteration at. This is 0 based. If not supplied assumes `0` |

**Input HTML**
```html
<html>
    <div>
        <table>
            <thead>
                <tr>
                    <th>name</th>
                    <th>age</th>
                    <th>link</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>Turing</td>
                    <td>41</td>
                    <td><a href="https://en.wikipedia.org/wiki/Alan_Turing">Wikipedia</a></td>
                </tr>
                <tr>
                    <td>Babbage</td>
                    <td>79</td>
                    <td><a href="https://en.wikipedia.org/wiki/Charles_Babbage">Wikipedia</a></td>
                </tr>
            </tbody>
        </table>
    </div>
</html>
```

**Input Schema**
```json
{
    "values": {
        "$type": "array",
        "$itr_el": "div.table.tbody",
        "$itr_ident": "table-row",
        "$itr_el_match": "tr",
        "$value_node": {
            "name": {
                "$type": "string",
                "$selector": "@table-row.td[nth=1]"
            },
            "age": {
                "$type": "string",
                "$selector": "@table-row.td[nth=1]"
            }
            "link": {
                "$type": "string",
                "$selector": "@table-row.td[nth=1]"
            }
        }
    }
}
```

**Output JSON**
```json
{
    "values": [
        {
            "name": "Turing",
            "age": "41",
            "link": "https://en.wikipedia.org/wiki/Alan_Turing"
        },
        {
            "name": "Babbage",
            "age": "79",
            "link": "https://en.wikipedia.org/wiki/Charles_Babbage"
        }
    ]
}
```


### `object`

| Field           | Required | Description |
|-----------------|----------|-------------|
| `$type`         | ✅       | `object`    |
| `$key_node`     | ✅       | The node to select for the object key. Can only be of type `string` or `number` node |
| `$value_node`   | ✅       | The node to select for the object value. Can be any type of node. |
| `$itr_el`       | ❌       | The node to iterate over to build the object. If not provided the object will only have one key value pair. |
| `$itr_ident`    | ❌       | The internal identifier to reference in the schema. This is required when `$itr_el` is supplied |
| `$itr_el_match` | ❌       | The tag to match against when iterating. This has the same power as the `$selector` field. This is required when `$itr_el` is supplied |

**Input HTML**
```html
<html>
    <div>
        <table>
            <tbody>
                <tr>
                    <td>name</td>
                    <td>Turing</td>
                </tr>
                <tr>
                    <td>age</td>
                    <td>41</td>
                </tr>
                <hr>
                <tr>
                    <td>link</td>
                    <td>https://en.wikipedia.org/wiki/Alan_Turing">Wikipedia</td>
                </tr>
            </tbody>
        </table>
    </div>
</html>
```

**Input Schema**
```json
{
    "singular": {
        "$type": "object",
        "$key_node": {
            "$type": "string",
            "$selector": "html.div.table.tbody.tr[!nth=0].td[!nth=0]"
        },
        "$value_node": {
            "$type": "string",
            "$selector": "html.div.table.tbody.tr[!nth=0].td[!nth=1]"
        }
    },
    "iteration": {
        "$type": "object",
        "$itr_el": "div.table.tbody",
        "$itr_ident": "table-row",
        "$itr_el_match": "tr",
        "$key_node": {
            "$type": "string",
            "$selector": "@table-row.td[nth=0]"
        },
        "$value_node": {
            "$type": "string",
            "$selector": "@table-row.td[nth=1]"
        }
    }
}
```

**Output JSON**
```json
{
    "singular": {
        "name": "Turing",
    },
    "iteration": {
        "name": "Turing",
        "age": "41",
        "link": "https://en.wikipedia.org/wiki/Alan_Turing"
    }
}
```

## Work in Progress

### Node
- [ ] Add support for `boolean` type
- [ ] Add support for `null` type
- [ ] Add support for !nth to have a value of `first`/`last`
- [ ] Add support for selecting out data via `:<attribute-name>`
- [ ] Add support for `?` (optional) access
- [ ] Add support for array with index access (useful for tables with headers)
