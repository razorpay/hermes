import RESTAdapter from "@ember-data/adapter/rest";
import { inject as service } from "@ember/service";
import DS from "ember-data";
import ModelRegistry from "ember-data/types/registries/model";
import ConfigService from "hermes/services/config";
import SessionService from "hermes/services/session";
import RSVP from "rsvp";

export default class ApplicationAdapter extends RESTAdapter {
  @service("config") declare configSvc: ConfigService;
  @service declare session: SessionService;

  namespace = "api/v1";

  get headers() {
    return {
      "Hermes-Google-Access-Token":
        this.session.data.authenticated.access_token,
    };
  }

  // findAll<K extends string | number>(
  //   store: DS.Store,
  //   type: ModelRegistry[K],
  //   sinceToken: string,
  //   snapshotRecordArray: DS.SnapshotRecordArray<K>
  // ): RSVP.Promise<any> {
  //   debugger;
  //   return super.findAll(store, type, sinceToken, snapshotRecordArray);
  // }
}
