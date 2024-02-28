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

- `!nth` - The index of the tag to select. This is 0 based or first/last. Possible values are `first`, `last`, or a number.

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

#### Get Data from Attributes `:<attribute-name>`

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

### String

| Field     | Required | Description |
|-----------|----------|-------------|
| $type     | ✅       | `string`    |
| $selector | ✅       | The node to select from the html. |

**Input HTML**
```html
<html>
    <div>
        <p>Turing</hjson>
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

### Number

**Input HTML**
```html
<html>
    <div>
        <p>Turing</hjson>
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
