// import JSONAPISerializer from "@ember-data/serializer/json-api";
import RESTSerializer from "@ember-data/serializer/rest";

export default class ApplicationSerializer extends RESTSerializer {
  normalizeResponse(
    store: any,
    primaryModelClass: any,
    payload: any,
    id: any,
    requestType: any
  ) {
    return super.normalizeResponse(
      store,
      primaryModelClass,
      payload,
      id,
      requestType
    );
  }
}
