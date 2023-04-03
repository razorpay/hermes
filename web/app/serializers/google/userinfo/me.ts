import RESTSerializer from "@ember-data/serializer/rest";

export default class GoogleDriveFileSerializer extends RESTSerializer {
  normalizeQueryRecordResponse(
    store: any,
    primaryModelClass: any,
    payload: any,
    id: any,
    requestType: any
  ) {
    return super.normalizeQueryResponse(
      store,
      primaryModelClass,
      {
        "google.userinfo.me": payload,
      },
      id,
      requestType
    );
  }
}
