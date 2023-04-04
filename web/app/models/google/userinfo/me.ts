import Model, { attr, hasMany } from "@ember-data/model";
import MeSubscriptionModel from "hermes/models/me/subscription";

export default class GoogleUserInfoMeModel extends Model {
  @attr() declare email: string;
  @attr() declare given_name: string;
  @attr() declare name: string;
  @attr() declare picture: string;

  @hasMany("me/subscription") declare subscriptions: MeSubscriptionModel[];
}
