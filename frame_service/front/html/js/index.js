var destination = document.querySelector("#container");

var { Router,
  Route,
  IndexRoute,
  IndexLink,
  Link } = ReactRouter;

var STUFFITEMS = [
{ id: 1, title: "Nulla pulvinar diam", visible: false },
{ id: 2, title: "Facilisis bibendum", visible: false },
{ id: 3, title: "Vestibulum vulputate", visible: false },
{ id: 4, title: "Eget erat", visible: false },
{ id: 5, title: "Id porttitor", visible: false },
{ id: 6, title: "Lorem Ipsum", visible: true }];


var App = React.createClass({ displayName: "App",
  render: function () {
    return (
      React.createElement("div", null,
      React.createElement("h1", null, "framecase.ru"),
      React.createElement("ul", { className: "header" },
      React.createElement("li", null, React.createElement(IndexLink, { to: "/", activeClassName: "active" }, "Начальная")),
      React.createElement("li", null, React.createElement(Link, { to: "/stuffs", activeClassName: "active" }, "Картинки")),
      React.createElement("li", null, React.createElement(Link, { to: "/faq", activeClassName: "active" }, "FAQ"))),

      React.createElement("div", { className: "content" },
      this.props.children)));



  } });


var Home = React.createClass({ displayName: "Home",
  render: function () {
    return (
      React.createElement("div", null,
      React.createElement("h2", null, "Здраствуй, друг!"),
      React.createElement("p", null, "Тестовый сайт"),
      React.createElement("p", null, "Frame2"),
      React.createElement("p", null, "A wrong link: ", React.createElement(Link, { to: "ding" }, "Yes please"), ".")));


  } });


var Faq = React.createClass({ displayName: "Faq",
  render: function () {
    return (
      React.createElement("div", null,
      React.createElement("h2", null, "Test"),
	  React.createElement("img",{src: `/v1/get/img`})));
      



  } });


var Stuffs = React.createClass({ displayName: "Stuffs",
  _renderStuffs: function () {
    return (
      React.createElement("div", null,
      React.createElement("h2", null, "STUFFS"),
      React.createElement("p", null, "Mauris sem velit, vehicula eget sodales vitae, rhoncus eget sapien:"),

      React.createElement("ol", null,
      this.state.stuffs.map((stuff) =>
      React.createElement("li", null, React.createElement(Link, { to: `/stuffs/${stuff.id}`, activeClassName: "active" }, stuff.title))))));




  },

  getInitialState: function () {
    return {
      stuffs: STUFFITEMS };

  },

  render: function () {
    return (
      React.createElement("div", null,
      this.props.params.stuffId ? this.props.children : this._renderStuffs()));


  } });


var Stuff = React.createClass({ displayName: "Stuff",
  _findStuffById: function (id) {
    return STUFFITEMS.filter(stuffItem => stuffItem.id == id)[0];
  },

  getInitialState: function () {
    return {
      stuffs: null };

  },

  componentWillMount: function () {
    this.setState({
      stuff: this._findStuffById(this.props.params.stuffId) });

  },

  render: function () {
    const stuff = this.state.stuff;
    return (
      React.createElement("div", null,
      React.createElement("h2", null, stuff.title),
      React.createElement("p", null, "My ID is ", stuff.id),
      React.createElement("p", null, React.createElement(Link, { to: "/stuffs" }, "Return to Stuffs"))));


  } });


var NoMatch = React.createClass({ displayName: "NoMatch",
  render: function () {
    return (
      React.createElement("div", null,
      React.createElement("h2", null, "No route matches this URL."),
      React.createElement("p", null, "Return ", React.createElement(Link, { to: "/" }, "home"))));


  } });


ReactDOM.render(
React.createElement(Router, null,
React.createElement(Route, { path: "/", component: App },
React.createElement(IndexRoute, { component: Home }),
React.createElement(Route, { path: "/stuffs", component: Stuffs },
React.createElement(Route, { path: ":stuffId", component: Stuff })),

React.createElement(Route, { path: "faq", component: Faq }),
React.createElement(Route, { path: "*", component: NoMatch }))),


destination);