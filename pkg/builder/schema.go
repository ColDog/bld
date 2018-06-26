package builder

const schema = `{"$schema":"http://json-schema.org/draft-06/schema#","title":"Build","type":"object","additionalProperties":false,"properties":{"id":{"type":"string"},"name":{"type":"string"},"volumes":{"type":"array","items":{"$ref":"#/definitions/volume"}},"sources":{"type":"array","items":{"$ref":"#/definitions/source"}},"steps":{"type":"array","items":{"$ref":"#/definitions/step"}}},"definitions":{"volume":{"type":"object","properties":{"name":{"type":"string"},"target":{"type":"string"}},"required":["name","target"]},"source":{"type":"object","properties":{"name":{"type":"string"},"target":{"type":"string"},"files":{"type":"array","items":{"type":"string"},"uniqueItems":true}},"required":["name","target"]},"mount":{"type":"object","properties":{"source":{"type":"string"},"mount":{"type":"string"}},"required":["source","mount"]},"image":{"type":"object","properties":{"tag":{"type":"string"},"entrypoint":{"type":"array","items":{"type":"string"}},"env":{"type":"array","items":{"type":"string"}},"workdir":{"type":"string"}},"required":["tag"]},"step":{"type":"object","properties":{"name":{"type":"string"},"image":{"type":"string"},"commands":{"type":"array","items":{"type":"string"}},"imports":{"type":"array","items":{"$ref":"#/definitions/mount"}},"exports":{"type":"array","items":{"$ref":"#/definitions/mount"}},"volumes":{"type":"array","items":{"$ref":"#/definitions/mount"}},"env":{"type":"array","items":{"type":"string"}},"workdir":{"type":"string"},"save":{"$ref":"#/definitions/image"}},"required":["name","image","commands"]}},"required":["name"]}`
