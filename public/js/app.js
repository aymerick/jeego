window.Jeego = Ember.Application.create();

Jeego.ApplicationAdapter = DS.RESTAdapter.extend({
  namespace: 'api'
});

// moment.lang('fr');
