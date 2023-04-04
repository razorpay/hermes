import JSONAPISerializer from "@ember-data/serializer/json-api";
import RESTSerializer from "@ember-data/serializer/rest";
// import Serializer from "@ember-data/serializer";
import { guidFor } from "@ember/object/internals";
import { SubscriptionType } from "hermes/services/authenticated-user";

export default class MeSubscriptionSerializer extends JSONAPISerializer {
  normalizeResponse(
    store: any,
    primaryModelClass: any,
    payload: any,
    id: any,
    requestType: any
  ) {
    let newPayload = payload.map((subscription: string) => {
      return {
        id: guidFor("me/subscription"),
        type: "me/subscription",
        attributes: {
          name: subscription,
          subscriptionType: SubscriptionType.Instant,
        },
      };
    });

    return { data: newPayload };
  }
}
