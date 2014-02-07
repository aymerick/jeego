// usage: {{dateFormat updated_at format="MMMM YYYY"}}
Ember.Handlebars.registerBoundHelper('dateFormat', function(value, options) {
  var f = options.hash.format || "LLL";
  return moment(value).format(f);
});

// usage: {{timeAgo updated_at}}
Ember.Handlebars.registerBoundHelper('timeAgo', function(value) {
  return moment(value).fromNow();
});
