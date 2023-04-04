import Model, { attr, belongsTo } from "@ember-data/model";
import { SubscriptionType } from "hermes/services/authenticated-user";
import GoogleUserInfoMeModel from "../google/userinfo/me";

export default class MeSubscriptionModel extends Model {
  @attr() declare name: string;
  @attr() declare subscriptionType: SubscriptionType;

  @belongsTo('google/userinfo/me') declare me: GoogleUserInfoMeModel;
}
