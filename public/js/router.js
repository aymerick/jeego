Jeego.Router.map(function() {
  this.resource('nodes', { path: '/' });
});

Jeego.NodesRoute = Ember.Route.extend({
  model: function() {
    return this.store.find('node');
  }
});
