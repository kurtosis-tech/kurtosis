const rootFolder = './api/protobuf/core/';
const fileNames = ['api_container_service.proto'];

module.exports = {
  files: fileNames.map(f => `${rootFolder}/${f}`),
  dist: `${rootFolder}/topic-specifications.schema.json`,
  long: 'number',
  transform(type, result, args) {
    switch (type) {
      case 'enum': {
        const [Enum] = args;
        // console.log('enum:', Enum);
        return { ...result, type: 'string', enum: Object.keys(Enum.values) };
      }
      case 'message': {
        const [Type] = args;
        // looking for Map fields and converting them to { [key: string]: type; }
        const map = Object.values(Type.fields || {}).filter(f => f.map && f.keyType === 'string');
        const newResult = { ...result };
        map.forEach(field => {
          newResult.properties[field.name].type = 'object';
          newResult.properties[field.name].additionalProperties = { type: 'string' };
        });

        // console.log('result', newResult);
        return newResult;
      }
    }
    return result;
  }
};

