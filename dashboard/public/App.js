'use strict';

function noop() { }
function run(fn) {
    return fn();
}
function blank_object() {
    return Object.create(null);
}
function run_all(fns) {
    fns.forEach(run);
}
function is_function(thing) {
    return typeof thing === 'function';
}
function safe_not_equal(a, b) {
    return a != a ? b == b : a !== b || ((a && typeof a === 'object') || typeof a === 'function');
}
function subscribe(store, ...callbacks) {
    if (store == null) {
        return noop;
    }
    const unsub = store.subscribe(...callbacks);
    return unsub.unsubscribe ? () => unsub.unsubscribe() : unsub;
}
function null_to_empty(value) {
    return value == null ? '' : value;
}

let current_component;
function set_current_component(component) {
    current_component = component;
}
function get_current_component() {
    if (!current_component)
        throw new Error('Function called outside component initialization');
    return current_component;
}
/**
 * The `onMount` function schedules a callback to run as soon as the component has been mounted to the DOM.
 * It must be called during the component's initialisation (but doesn't need to live *inside* the component;
 * it can be called from an external module).
 *
 * `onMount` does not run inside a [server-side component](/docs#run-time-server-side-component-api).
 *
 * https://svelte.dev/docs#run-time-svelte-onmount
 */
function onMount(fn) {
    get_current_component().$$.on_mount.push(fn);
}
/**
 * Schedules a callback to run immediately before the component is unmounted.
 *
 * Out of `onMount`, `beforeUpdate`, `afterUpdate` and `onDestroy`, this is the
 * only one that runs inside a server-side component.
 *
 * https://svelte.dev/docs#run-time-svelte-ondestroy
 */
function onDestroy(fn) {
    get_current_component().$$.on_destroy.push(fn);
}
/**
 * Associates an arbitrary `context` object with the current component and the specified `key`
 * and returns that object. The context is then available to children of the component
 * (including slotted content) with `getContext`.
 *
 * Like lifecycle functions, this must be called during component initialisation.
 *
 * https://svelte.dev/docs#run-time-svelte-setcontext
 */
function setContext(key, context) {
    get_current_component().$$.context.set(key, context);
    return context;
}
/**
 * Retrieves the context that belongs to the closest parent component with the specified `key`.
 * Must be called during component initialisation.
 *
 * https://svelte.dev/docs#run-time-svelte-getcontext
 */
function getContext(key) {
    return get_current_component().$$.context.get(key);
}
Promise.resolve();
const ATTR_REGEX = /[&"]/g;
const CONTENT_REGEX = /[&<]/g;
/**
 * Note: this method is performance sensitive and has been optimized
 * https://github.com/sveltejs/svelte/pull/5701
 */
function escape(value, is_attr = false) {
    const str = String(value);
    const pattern = is_attr ? ATTR_REGEX : CONTENT_REGEX;
    pattern.lastIndex = 0;
    let escaped = '';
    let last = 0;
    while (pattern.test(str)) {
        const i = pattern.lastIndex - 1;
        const ch = str[i];
        escaped += str.substring(last, i) + (ch === '&' ? '&amp;' : (ch === '"' ? '&quot;' : '&lt;'));
        last = i + 1;
    }
    return escaped + str.substring(last);
}
function each(items, fn) {
    let str = '';
    for (let i = 0; i < items.length; i += 1) {
        str += fn(items[i], i);
    }
    return str;
}
const missing_component = {
    $$render: () => ''
};
function validate_component(component, name) {
    if (!component || !component.$$render) {
        if (name === 'svelte:component')
            name += ' this={...}';
        throw new Error(`<${name}> is not a valid SSR component. You may need to review your build config to ensure that dependencies are compiled, rather than imported as pre-compiled modules. Otherwise you may need to fix a <${name}>.`);
    }
    return component;
}
let on_destroy;
function create_ssr_component(fn) {
    function $$render(result, props, bindings, slots, context) {
        const parent_component = current_component;
        const $$ = {
            on_destroy,
            context: new Map(context || (parent_component ? parent_component.$$.context : [])),
            // these will be immediately discarded
            on_mount: [],
            before_update: [],
            after_update: [],
            callbacks: blank_object()
        };
        set_current_component({ $$ });
        const html = fn(result, props, bindings, slots);
        set_current_component(parent_component);
        return html;
    }
    return {
        render: (props = {}, { $$slots = {}, context = new Map() } = {}) => {
            on_destroy = [];
            const result = { title: '', head: '', css: new Set() };
            const html = $$render(result, props, {}, $$slots, context);
            run_all(on_destroy);
            return {
                html,
                css: {
                    code: Array.from(result.css).map(css => css.code).join('\n'),
                    map: null // TODO
                },
                head: result.title + result.head
            };
        },
        $$render
    };
}
function add_attribute(name, value, boolean) {
    if (value == null || (boolean && !value))
        return '';
    const assignment = (boolean && value === true) ? '' : `="${escape(value, true)}"`;
    return ` ${name}${assignment}`;
}

const subscriber_queue = [];
/**
 * Creates a `Readable` store that allows reading by subscription.
 * @param value initial value
 * @param {StartStopNotifier}start start and stop notifications for subscriptions
 */
function readable(value, start) {
    return {
        subscribe: writable(value, start).subscribe
    };
}
/**
 * Create a `Writable` store that allows both updating and reading by subscription.
 * @param {*=}value initial value
 * @param {StartStopNotifier=}start start and stop notifications for subscriptions
 */
function writable(value, start = noop) {
    let stop;
    const subscribers = new Set();
    function set(new_value) {
        if (safe_not_equal(value, new_value)) {
            value = new_value;
            if (stop) { // store is ready
                const run_queue = !subscriber_queue.length;
                for (const subscriber of subscribers) {
                    subscriber[1]();
                    subscriber_queue.push(subscriber, value);
                }
                if (run_queue) {
                    for (let i = 0; i < subscriber_queue.length; i += 2) {
                        subscriber_queue[i][0](subscriber_queue[i + 1]);
                    }
                    subscriber_queue.length = 0;
                }
            }
        }
    }
    function update(fn) {
        set(fn(value));
    }
    function subscribe(run, invalidate = noop) {
        const subscriber = [run, invalidate];
        subscribers.add(subscriber);
        if (subscribers.size === 1) {
            stop = start(set) || noop;
        }
        run(value);
        return () => {
            subscribers.delete(subscriber);
            if (subscribers.size === 0) {
                stop();
                stop = null;
            }
        };
    }
    return { set, update, subscribe };
}
function derived(stores, fn, initial_value) {
    const single = !Array.isArray(stores);
    const stores_array = single
        ? [stores]
        : stores;
    const auto = fn.length < 2;
    return readable(initial_value, (set) => {
        let inited = false;
        const values = [];
        let pending = 0;
        let cleanup = noop;
        const sync = () => {
            if (pending) {
                return;
            }
            cleanup();
            const result = fn(single ? values[0] : values, set);
            if (auto) {
                set(result);
            }
            else {
                cleanup = is_function(result) ? result : noop;
            }
        };
        const unsubscribers = stores_array.map((store, i) => subscribe(store, (value) => {
            values[i] = value;
            pending &= ~(1 << i);
            if (inited) {
                sync();
            }
        }, () => {
            pending |= (1 << i);
        }));
        inited = true;
        sync();
        return function stop() {
            run_all(unsubscribers);
            cleanup();
        };
    });
}

const LOCATION = {};
const ROUTER = {};

/**
 * Adapted from https://github.com/reach/router/blob/b60e6dd781d5d3a4bdaaf4de665649c0f6a7e78d/src/lib/history.js
 *
 * https://github.com/reach/router/blob/master/LICENSE
 * */

function getLocation(source) {
  return {
    ...source.location,
    state: source.history.state,
    key: (source.history.state && source.history.state.key) || "initial"
  };
}

function createHistory(source, options) {
  const listeners = [];
  let location = getLocation(source);

  return {
    get location() {
      return location;
    },

    listen(listener) {
      listeners.push(listener);

      const popstateListener = () => {
        location = getLocation(source);
        listener({ location, action: "POP" });
      };

      source.addEventListener("popstate", popstateListener);

      return () => {
        source.removeEventListener("popstate", popstateListener);

        const index = listeners.indexOf(listener);
        listeners.splice(index, 1);
      };
    },

    navigate(to, { state, replace = false } = {}) {
      state = { ...state, key: Date.now() + "" };
      // try...catch iOS Safari limits to 100 pushState calls
      try {
        if (replace) {
          source.history.replaceState(state, null, to);
        } else {
          source.history.pushState(state, null, to);
        }
      } catch (e) {
        source.location[replace ? "replace" : "assign"](to);
      }

      location = getLocation(source);
      listeners.forEach(listener => listener({ location, action: "PUSH" }));
    }
  };
}

// Stores history entries in memory for testing or other platforms like Native
function createMemorySource(initialPathname = "/") {
  let index = 0;
  const stack = [{ pathname: initialPathname, search: "" }];
  const states = [];

  return {
    get location() {
      return stack[index];
    },
    addEventListener(name, fn) {},
    removeEventListener(name, fn) {},
    history: {
      get entries() {
        return stack;
      },
      get index() {
        return index;
      },
      get state() {
        return states[index];
      },
      pushState(state, _, uri) {
        const [pathname, search = ""] = uri.split("?");
        index++;
        stack.push({ pathname, search });
        states.push(state);
      },
      replaceState(state, _, uri) {
        const [pathname, search = ""] = uri.split("?");
        stack[index] = { pathname, search };
        states[index] = state;
      }
    }
  };
}

// Global history uses window.history as the source if available,
// otherwise a memory history
const canUseDOM = Boolean(
  typeof window !== "undefined" &&
    window.document &&
    window.document.createElement
);
const globalHistory = createHistory(canUseDOM ? window : createMemorySource());

/**
 * Adapted from https://github.com/reach/router/blob/b60e6dd781d5d3a4bdaaf4de665649c0f6a7e78d/src/lib/utils.js
 *
 * https://github.com/reach/router/blob/master/LICENSE
 * */

const paramRe = /^:(.+)/;

const SEGMENT_POINTS = 4;
const STATIC_POINTS = 3;
const DYNAMIC_POINTS = 2;
const SPLAT_PENALTY = 1;
const ROOT_POINTS = 1;

/**
 * Check if `segment` is a root segment
 * @param {string} segment
 * @return {boolean}
 */
function isRootSegment(segment) {
  return segment === "";
}

/**
 * Check if `segment` is a dynamic segment
 * @param {string} segment
 * @return {boolean}
 */
function isDynamic(segment) {
  return paramRe.test(segment);
}

/**
 * Check if `segment` is a splat
 * @param {string} segment
 * @return {boolean}
 */
function isSplat(segment) {
  return segment[0] === "*";
}

/**
 * Split up the URI into segments delimited by `/`
 * @param {string} uri
 * @return {string[]}
 */
function segmentize(uri) {
  return (
    uri
      // Strip starting/ending `/`
      .replace(/(^\/+|\/+$)/g, "")
      .split("/")
  );
}

/**
 * Strip `str` of potential start and end `/`
 * @param {string} str
 * @return {string}
 */
function stripSlashes(str) {
  return str.replace(/(^\/+|\/+$)/g, "");
}

/**
 * Score a route depending on how its individual segments look
 * @param {object} route
 * @param {number} index
 * @return {object}
 */
function rankRoute(route, index) {
  const score = route.default
    ? 0
    : segmentize(route.path).reduce((score, segment) => {
        score += SEGMENT_POINTS;

        if (isRootSegment(segment)) {
          score += ROOT_POINTS;
        } else if (isDynamic(segment)) {
          score += DYNAMIC_POINTS;
        } else if (isSplat(segment)) {
          score -= SEGMENT_POINTS + SPLAT_PENALTY;
        } else {
          score += STATIC_POINTS;
        }

        return score;
      }, 0);

  return { route, score, index };
}

/**
 * Give a score to all routes and sort them on that
 * @param {object[]} routes
 * @return {object[]}
 */
function rankRoutes(routes) {
  return (
    routes
      .map(rankRoute)
      // If two routes have the exact same score, we go by index instead
      .sort((a, b) =>
        a.score < b.score ? 1 : a.score > b.score ? -1 : a.index - b.index
      )
  );
}

/**
 * Ranks and picks the best route to match. Each segment gets the highest
 * amount of points, then the type of segment gets an additional amount of
 * points where
 *
 *  static > dynamic > splat > root
 *
 * This way we don't have to worry about the order of our routes, let the
 * computers do it.
 *
 * A route looks like this
 *
 *  { path, default, value }
 *
 * And a returned match looks like:
 *
 *  { route, params, uri }
 *
 * @param {object[]} routes
 * @param {string} uri
 * @return {?object}
 */
function pick(routes, uri) {
  let match;
  let default_;

  const [uriPathname] = uri.split("?");
  const uriSegments = segmentize(uriPathname);
  const isRootUri = uriSegments[0] === "";
  const ranked = rankRoutes(routes);

  for (let i = 0, l = ranked.length; i < l; i++) {
    const route = ranked[i].route;
    let missed = false;

    if (route.default) {
      default_ = {
        route,
        params: {},
        uri
      };
      continue;
    }

    const routeSegments = segmentize(route.path);
    const params = {};
    const max = Math.max(uriSegments.length, routeSegments.length);
    let index = 0;

    for (; index < max; index++) {
      const routeSegment = routeSegments[index];
      const uriSegment = uriSegments[index];

      if (routeSegment !== undefined && isSplat(routeSegment)) {
        // Hit a splat, just grab the rest, and return a match
        // uri:   /files/documents/work
        // route: /files/* or /files/*splatname
        const splatName = routeSegment === "*" ? "*" : routeSegment.slice(1);

        params[splatName] = uriSegments
          .slice(index)
          .map(decodeURIComponent)
          .join("/");
        break;
      }

      if (uriSegment === undefined) {
        // URI is shorter than the route, no match
        // uri:   /users
        // route: /users/:userId
        missed = true;
        break;
      }

      let dynamicMatch = paramRe.exec(routeSegment);

      if (dynamicMatch && !isRootUri) {
        const value = decodeURIComponent(uriSegment);
        params[dynamicMatch[1]] = value;
      } else if (routeSegment !== uriSegment) {
        // Current segments don't match, not dynamic, not splat, so no match
        // uri:   /users/123/settings
        // route: /users/:id/profile
        missed = true;
        break;
      }
    }

    if (!missed) {
      match = {
        route,
        params,
        uri: "/" + uriSegments.slice(0, index).join("/")
      };
      break;
    }
  }

  return match || default_ || null;
}

/**
 * Check if the `path` matches the `uri`.
 * @param {string} path
 * @param {string} uri
 * @return {?object}
 */
function match(route, uri) {
  return pick([route], uri);
}

/**
 * Combines the `basepath` and the `path` into one path.
 * @param {string} basepath
 * @param {string} path
 */
function combinePaths(basepath, path) {
  return `${stripSlashes(
    path === "/" ? basepath : `${stripSlashes(basepath)}/${stripSlashes(path)}`
  )}/`;
}

/* node_modules\svelte-routing\src\Router.svelte generated by Svelte v3.53.1 */

const Router = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let $location, $$unsubscribe_location;
	let $routes, $$unsubscribe_routes;
	let $base, $$unsubscribe_base;
	let { basepath = "/" } = $$props;
	let { url = null } = $$props;
	const locationContext = getContext(LOCATION);
	const routerContext = getContext(ROUTER);
	const routes = writable([]);
	$$unsubscribe_routes = subscribe(routes, value => $routes = value);
	const activeRoute = writable(null);
	let hasActiveRoute = false; // Used in SSR to synchronously set that a Route is active.

	// If locationContext is not set, this is the topmost Router in the tree.
	// If the `url` prop is given we force the location to it.
	const location = locationContext || writable(url ? { pathname: url } : globalHistory.location);

	$$unsubscribe_location = subscribe(location, value => $location = value);

	// If routerContext is set, the routerBase of the parent Router
	// will be the base for this Router's descendants.
	// If routerContext is not set, the path and resolved uri will both
	// have the value of the basepath prop.
	const base = routerContext
	? routerContext.routerBase
	: writable({ path: basepath, uri: basepath });

	$$unsubscribe_base = subscribe(base, value => $base = value);

	const routerBase = derived([base, activeRoute], ([base, activeRoute]) => {
		// If there is no activeRoute, the routerBase will be identical to the base.
		if (activeRoute === null) {
			return base;
		}

		const { path: basepath } = base;
		const { route, uri } = activeRoute;

		// Remove the potential /* or /*splatname from
		// the end of the child Routes relative paths.
		const path = route.default
		? basepath
		: route.path.replace(/\*.*$/, "");

		return { path, uri };
	});

	function registerRoute(route) {
		const { path: basepath } = $base;
		let { path } = route;

		// We store the original path in the _path property so we can reuse
		// it when the basepath changes. The only thing that matters is that
		// the route reference is intact, so mutation is fine.
		route._path = path;

		route.path = combinePaths(basepath, path);

		if (typeof window === "undefined") {
			// In SSR we should set the activeRoute immediately if it is a match.
			// If there are more Routes being registered after a match is found,
			// we just skip them.
			if (hasActiveRoute) {
				return;
			}

			const matchingRoute = match(route, $location.pathname);

			if (matchingRoute) {
				activeRoute.set(matchingRoute);
				hasActiveRoute = true;
			}
		} else {
			routes.update(rs => {
				rs.push(route);
				return rs;
			});
		}
	}

	function unregisterRoute(route) {
		routes.update(rs => {
			const index = rs.indexOf(route);
			rs.splice(index, 1);
			return rs;
		});
	}

	if (!locationContext) {
		// The topmost Router in the tree is responsible for updating
		// the location store and supplying it through context.
		onMount(() => {
			const unlisten = globalHistory.listen(history => {
				location.set(history.location);
			});

			return unlisten;
		});

		setContext(LOCATION, location);
	}

	setContext(ROUTER, {
		activeRoute,
		base,
		routerBase,
		registerRoute,
		unregisterRoute
	});

	if ($$props.basepath === void 0 && $$bindings.basepath && basepath !== void 0) $$bindings.basepath(basepath);
	if ($$props.url === void 0 && $$bindings.url && url !== void 0) $$bindings.url(url);

	{
		{
			const { path: basepath } = $base;

			routes.update(rs => {
				rs.forEach(r => r.path = combinePaths(basepath, r._path));
				return rs;
			});
		}
	}

	{
		{
			const bestMatch = pick($routes, $location.pathname);
			activeRoute.set(bestMatch);
		}
	}

	$$unsubscribe_location();
	$$unsubscribe_routes();
	$$unsubscribe_base();
	return `${slots.default ? slots.default({}) : ``}`;
});

/* node_modules\svelte-routing\src\Route.svelte generated by Svelte v3.53.1 */

const Route = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let $activeRoute, $$unsubscribe_activeRoute;
	let $location, $$unsubscribe_location;
	let { path = "" } = $$props;
	let { component = null } = $$props;
	const { registerRoute, unregisterRoute, activeRoute } = getContext(ROUTER);
	$$unsubscribe_activeRoute = subscribe(activeRoute, value => $activeRoute = value);
	const location = getContext(LOCATION);
	$$unsubscribe_location = subscribe(location, value => $location = value);

	const route = {
		path,
		// If no path prop is given, this Route will act as the default Route
		// that is rendered if no other Route in the Router is a match.
		default: path === ""
	};

	let routeParams = {};
	let routeProps = {};
	registerRoute(route);

	// There is no need to unregister Routes in SSR since it will all be
	// thrown away anyway.
	if (typeof window !== "undefined") {
		onDestroy(() => {
			unregisterRoute(route);
		});
	}

	if ($$props.path === void 0 && $$bindings.path && path !== void 0) $$bindings.path(path);
	if ($$props.component === void 0 && $$bindings.component && component !== void 0) $$bindings.component(component);

	{
		if ($activeRoute && $activeRoute.route === route) {
			routeParams = $activeRoute.params;
		}
	}

	{
		{
			const { path, component, ...rest } = $$props;
			routeProps = rest;
		}
	}

	$$unsubscribe_activeRoute();
	$$unsubscribe_location();

	return `${$activeRoute !== null && $activeRoute.route === route
	? `${component !== null
		? `${validate_component(component || missing_component, "svelte:component").$$render($$result, Object.assign({ location: $location }, routeParams, routeProps), {}, {})}`
		: `${slots.default
			? slots.default({ params: routeParams, location: $location })
			: ``}`}`
	: ``}`;
});

// import hljs from "highlight.js";
// import python from 'highlight.js/lib/languages/python';
// hljs.registerLanguage('python', python);
let frameworkExamples = {
    Django: {
        install: "pip install api-analytics",
        codeFile: 'settings.py',
        example: `ANALYTICS_API_KEY = <api_key>

MIDDLEWARE = [
    'api_analytics.django.Analytics',
    ...
]`
    },
    Flask: {
        install: "pip install api-analytics",
        codeFile: '',
        example: `from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, api_key=<api_key>)  # Add middleware

@app.get('/')
async def root():
    return {'message': 'Hello World!'}`
    },
    FastAPI: {
        install: "pip install api-analytics",
        codeFile: '',
        example: `from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <api_key>)  # Add middleware

@app.get('/')
def root():
    return {'message': 'Hello World!'}`
    },
    Tornado: {
        install: "pip install api-analytics",
        codeFile: '',
        example: `import asyncio
from tornado.web import Application

from api_analytics.tornado import Analytics

# Inherit from the Analytics middleware class
class MainHandler(Analytics):
    def __init__(self, app, res):
        super().__init__(app, res, <api_key>)  # Pass api key

    def get(self):
        self.write({'message': 'Hello World!'})

def make_app():
    return Application([
        (r"/", MainHandler),
    ])

async def main():
    app = make_app()
    app.listen(8080)
    await asyncio.Event().wait()

if __name__ == "__main__":
    asyncio.run(main())`
    },
    Express: {
        install: 'npm install node-api-analytics',
        codeFile: '',
        example: `import express from 'express';
import { expressAnalytics } from 'node-api-analytics';

const app = express();

app.use(analytics(<api_key>));  // Add middleware

app.get('/', (req, res) => {
    res.send({ message: 'Hello World' });
});

app.listen(8080, () => {
    console.log('Server listening at http://localhost:8080');
})`
    },
    Fastify: {
        install: 'npm install node-api-analytics',
        codeFile: '',
        example: `import Fastify from 'fastify';
import { fastifyAnalytics } from 'node-api-analytics;

const fastify = Fastify();

fastify.addHook('onRequest', fastifyAnalytics(<api_key>));  // Add middleware

fastify.get('/', function (request, reply) {
  reply.send({ message: 'Hello World!' });
})

fastify.listen({ port: 8080 }, function (err, address) {
  console.log('Server listening at http://localhost:8080');
  if (err) {
    fastify.log.error(err);
    process.exit(1);
  }
})`
    },
    Koa: {
        install: 'npm install node-api-analytics',
        codeFile: '',
        example: `import Koa from "koa";
import { koaAnalytics } from 'node-api-analytics';

const app = new Koa();

app.use(koaAnalytics(<api_key>));  // Add middleware

app.use((ctx) => {
  ctx.body = { message: 'Hello World!' };
});

app.listen(8080, () =>
  console.log('Server listening at https://localhost:8080')
);`
    },
    Gin: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/gin',
        codeFile: '',
        example: `package main

import (
	analytics "github.com/tom-draper/api-analytics/analytics/go/gin"
	"net/http"

	"github.com/gin-gonic/gin"
)

func root(c *gin.Context) {
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	router := gin.Default()
	
	router.Use(analytics.Analytics(<api_key>)) // Add middleware

	router.GET("/", root)
	router.Run("localhost:8080")
}`
    },
    Echo: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/echo',
        codeFile: '',
        example: `package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	analytics "github.com/tom-draper/api-analytics/analytics/go/echo"
)

func root(c echo.Context) {
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	router := echo.New()

	router.Use(analytics.Analytics(<api_key>)) // Add middleware

	router.GET("/", root)
	router.Start("localhost:8080")
}`
    },
    Fiber: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/fiber',
        codeFile: '',
        example: `package main

import (
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/fiber"

	"github.com/gofiber/fiber/v2"
)

func root(c *fiber.Ctx) error {
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	return c.SendString(string(jsonData))
}

func main() {
	app := fiber.New()

	app.Use(analytics.Analytics(<api_key>)) // Add middleware

	app.Get("/", root)
	app.Listen(":8080")
}`
    },
    Chi: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/chi',
        codeFile: '',
        example: `package main

import (
	"net/http"
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/chi"

	chi "github.com/go-chi/chi/v5"
)

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	w.Write(jsonData)
}

func main() {
	router := chi.NewRouter()

	router.Use(analytics.Analytics(<api_key>)) // Add middleware

	router.GET("/", root)
	router.Run("localhost:8080")
}`
    },
    Actix: {
        install: 'cargo add actix-analytics',
        codeFile: '',
        example: `use actix_web::{get, web, Responder, Result};
use serde::Serialize;
use actix_analytics::Analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

#[get("/")]
async fn index() -> Result<impl Responder> {
    let json_data = JsonData {
        message: "Hello World!".to_string(),
    };
    Ok(web::Json(json_data))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    use actix_web::{App, HttpServer};

    HttpServer::new(|| {
        App::new()
            .wrap(Analytics::new(<api_key>))  // Add middleware
            .service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}`
    },
    Axum: {
        install: 'cargo add axum-analytics',
        codeFile: '',
        example: `use axum::{
    routing::get,
    Json, Router,
};
use serde::Serialize;
use std::net::SocketAddr;
use tokio;
use actum_analytics::Analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

async fn root() -> Json<JsonData> {
    let json_data = JsonData {
        message: "Hello World!".to_string(),
    };
    Json(json_data)
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .layer(Analytics::new(<api_key>))  // Add middleware
        .route("/", get(root));

    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}`
    },
    Rails: {
        install: 'gem install api_analytics',
        codeFile: 'config/application.rb',
        example: `require 'rails'
require 'api_analytics'

Bundler.require(*Rails.groups)

module RailsMiddleware
  class Application < Rails::Application
    config.load_defaults 6.1
    config.api_only = true

    config.middleware.use ::Analytics::Rails, <api_key> # Add middleware
  end
end`
    },
    Sinatra: {
        install: 'gem install api_analytics',
        codeFile: '',
        example: `require 'sinatra'
require 'api_analytics'

use Analytics::Sinatra, <api_key>

before do
    content_type 'application/json'
end

get '/' do
    {message: 'Hello World!'}.to_json
end`
    }
};

/* src\components\Footer.svelte generated by Svelte v3.53.1 */

const css$n = {
	code: ".logo.svelte-82t1ww{font-size:0.9em;color:var(--highlight)}.footer.svelte-82t1ww{margin:1.5em 0 3em}.github-logo.svelte-82t1ww{height:30px;filter:invert(0.24);margin-bottom:30px}",
	map: "{\"version\":3,\"file\":\"Footer.svelte\",\"sources\":[\"Footer.svelte\"],\"sourcesContent\":[\"<div class=\\\"footer\\\">\\r\\n  <a class=\\\"github-link\\\" rel=\\\"noreferrer\\\" target=\\\"_blank\\\" href=\\\"https://github.com/tom-draper/api-analytics\\\">\\r\\n    <img class=\\\"github-logo\\\" height=\\\"30px\\\" src=\\\"../img/github.png\\\" alt=\\\"\\\" />\\r\\n  </a>\\r\\n  <div class=\\\"logo\\\">API Analytics</div>\\r\\n  <img class=\\\"footer-logo\\\" src=\\\"../img/logo.png\\\" alt=\\\"\\\" />\\r\\n</div>\\r\\n\\r\\n<style scoped>\\r\\n  .logo {\\r\\n    font-size: 0.9em;\\r\\n    color: var(--highlight);\\r\\n  }\\r\\n  .footer {\\r\\n    margin: 1.5em 0 3em;\\r\\n  }\\r\\n  .github-logo {\\r\\n    height: 30px;\\r\\n    filter: invert(0.24);\\r\\n    margin-bottom: 30px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AASE,KAAK,cAAC,CAAC,AACL,SAAS,CAAE,KAAK,CAChB,KAAK,CAAE,IAAI,WAAW,CAAC,AACzB,CAAC,AACD,OAAO,cAAC,CAAC,AACP,MAAM,CAAE,KAAK,CAAC,CAAC,CAAC,GAAG,AACrB,CAAC,AACD,YAAY,cAAC,CAAC,AACZ,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,OAAO,IAAI,CAAC,CACpB,aAAa,CAAE,IAAI,AACrB,CAAC\"}"
};

const Footer = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	$$result.css.add(css$n);

	return `<div class="${"footer svelte-82t1ww"}"><a class="${"github-link"}" rel="${"noreferrer"}" target="${"_blank"}" href="${"https://github.com/tom-draper/api-analytics"}"><img class="${"github-logo svelte-82t1ww"}" height="${"30px"}" src="${"../img/github.png"}" alt="${""}"></a>
  <div class="${"logo svelte-82t1ww"}">API Analytics</div>
  <img class="${"footer-logo"}" src="${"../img/logo.png"}" alt="${""}">
</div>`;
});

const a11yDark = `<style>pre code.hljs{display:block;overflow-x:auto;padding:1em}code.hljs{padding:3px 5px}/*!
  Theme: a11y-dark
  Author: @ericwbailey
  Maintainer: @ericwbailey

  Based on the Tomorrow Night Eighties theme: https://github.com/isagalaev/highlight.js/blob/master/src/styles/tomorrow-night-eighties.css
*/.hljs{background:#2b2b2b;color:#f8f8f2}.hljs-comment,.hljs-quote{color:#d4d0ab}.hljs-deletion,.hljs-name,.hljs-regexp,.hljs-selector-class,.hljs-selector-id,.hljs-tag,.hljs-template-variable,.hljs-variable{color:#ffa07a}.hljs-built_in,.hljs-link,.hljs-literal,.hljs-meta,.hljs-number,.hljs-params,.hljs-type{color:#f5ab35}.hljs-attribute{color:gold}.hljs-addition,.hljs-bullet,.hljs-string,.hljs-symbol{color:#abe338}.hljs-section,.hljs-title{color:#00e0e0}.hljs-keyword,.hljs-selector-tag{color:#dcc6e0}.hljs-emphasis{font-style:italic}.hljs-strong{font-weight:700}@media screen and (-ms-high-contrast:active){.hljs-addition,.hljs-attribute,.hljs-built_in,.hljs-bullet,.hljs-comment,.hljs-link,.hljs-literal,.hljs-meta,.hljs-number,.hljs-params,.hljs-quote,.hljs-string,.hljs-symbol,.hljs-type{color:highlight}.hljs-keyword,.hljs-selector-tag{font-weight:700}}</style>`;

/* src\routes\Home.svelte generated by Svelte v3.53.1 */

const css$m = {
	code: ".landing-page.svelte-qzx5tq{display:flex;place-items:center;padding-bottom:6em}.landing-page-container.svelte-qzx5tq{display:grid;margin:0 12% 0 15%;min-height:100vh}.left.svelte-qzx5tq{flex:1;text-align:left}.right.svelte-qzx5tq{display:grid;place-items:center}.logo.svelte-qzx5tq{max-width:1400px;width:700px;margin-bottom:-50px}h1.svelte-qzx5tq{font-size:3.4em}h2.svelte-qzx5tq{color:white}.links.svelte-qzx5tq{color:#707070;display:flex;margin-top:30px;text-align:left}.link.svelte-qzx5tq{width:fit-content;margin-right:20px;font-size:0.9}.link.svelte-qzx5tq:hover{background:#31aa73}.secondary.svelte-qzx5tq{background:#1c1c1c;border:3px solid var(--highlight);color:var(--highlight)}.secondary.svelte-qzx5tq:hover{background:#081d13}.lightning-top.svelte-qzx5tq{height:70px}.dashboard-title-container.svelte-qzx5tq{display:flex;margin:2em 4em}a.svelte-qzx5tq{background:var(--highlight);color:black;padding:10px 20px;border-radius:4px}.italic.svelte-qzx5tq{font-style:italic}.dashboard.svelte-qzx5tq{border:3px solid var(--highlight);width:80%;border-radius:10px;margin:auto;margin-bottom:8em;overflow:hidden;position:relative}.dashboard-title.svelte-qzx5tq{font-size:2.5em;text-align:left;margin:0.2em 1em auto 1em}.dashboard-title-container.svelte-qzx5tq{place-content:center}.dashboard-img.svelte-qzx5tq{width:81%;border-radius:10px;box-shadow:0px 24px 120px -25px var(--highlight);margin-bottom:-1%}.dashboard-content.svelte-qzx5tq{text-align:left;color:white;margin:0 5em 4em}.dashboard-content-text.svelte-qzx5tq{font-size:1.1em;text-align:center}.dashboard-btn-container.svelte-qzx5tq{display:flex;justify-content:center;margin-top:2em}.dashboard-btn-text.svelte-qzx5tq{text-align:center}.add-middleware-title.svelte-qzx5tq{color:var(--highlight);font-size:2.1em;font-weight:700;margin-bottom:1.5em}.add-middleware.svelte-qzx5tq{margin:auto;margin-bottom:7em;border-radius:6px}.frameworks.svelte-qzx5tq{margin:0 10%;overflow-x:auto}.framework.svelte-qzx5tq{color:grey;background:transparent;border:none;font-size:1em;cursor:pointer;padding:10px 18px;border-bottom:3px solid transparent}.active.svelte-qzx5tq{color:white}.active.python.svelte-qzx5tq{border-bottom:3px solid #4b8bbe}.active.golang.svelte-qzx5tq{border-bottom:3px solid #00a7d0}.active.javascript.svelte-qzx5tq{border-bottom:3px solid #edd718}.active.rust.svelte-qzx5tq{border-bottom:3px solid #ef4900}.active.ruby.svelte-qzx5tq{border-bottom:3px solid #cd0000}.subtitle.svelte-qzx5tq{color:rgb(110, 110, 110);margin:10px 0 2px 20px;font-size:0.85em}.instructions-container.svelte-qzx5tq{padding:1.5em 2em 2em;width:850px;margin:auto}.instructions.svelte-qzx5tq{text-align:left;display:flex;flex-direction:column;position:relative}code.svelte-qzx5tq{background:#151515;padding:1.4em 2em;border-radius:0.5em;margin:5px;color:#dcdfe4;white-space:pre-wrap}.code.svelte-qzx5tq{display:none}.code-file.svelte-qzx5tq{position:absolute;font-size:0.8em;top:160px;color:rgb(97, 97, 97);text-align:right;right:2.5em;margin-bottom:-2em}#hover-1.svelte-qzx5tq{transform:translateY(30px)}#hover-2.svelte-qzx5tq{transform:translateY(-30px)}img.svelte-qzx5tq{transition:all 10s ease-in-out}@media screen and (max-width: 1500px){.landing-page-container.svelte-qzx5tq{margin:0 6% 0 7%}}@media screen and (max-width: 1300px){.landing-page-container.svelte-qzx5tq{margin:0 5% 0 6%}}@media screen and (max-width: 1200px){.landing-page.svelte-qzx5tq{flex-direction:column-reverse}.landing-page-container.svelte-qzx5tq{margin:0 2em}.dashboard.svelte-qzx5tq{width:90%;margin-bottom:4em}.logo.svelte-qzx5tq{width:100%}}@media screen and (max-width: 900px){.home.svelte-qzx5tq{font-size:0.85em}.instructions-container.svelte-qzx5tq{width:auto}.code-file.svelte-qzx5tq{top:140px}}@media screen and (max-width: 800px){.home.svelte-qzx5tq{font-size:0.8em}.right.svelte-qzx5tq{margin-top:2em}}@media screen and (max-width: 700px){h1.svelte-qzx5tq{font-size:2.5em}h2.svelte-qzx5tq{font-size:1.2em}.landing-page-container.svelte-qzx5tq{min-height:unset}.landing-page.svelte-qzx5tq{padding-bottom:8em}.lightning-top.svelte-qzx5tq{height:55px}.dashboard-img.svelte-qzx5tq{margin-bottom:-3%}.add-middleware-title.svelte-qzx5tq{margin-bottom:0.5em}.instructions-container.svelte-qzx5tq{padding-top:0}}",
	map: "{\"version\":3,\"file\":\"Home.svelte\",\"sources\":[\"Home.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import frameworkExamples from \\\"../lib/framework\\\";\\r\\nimport Footer from \\\"../components/Footer.svelte\\\";\\r\\nimport codeStyle from \\\"svelte-highlight/styles/a11y-dark\\\";\\r\\nimport { onMount } from \\\"svelte\\\";\\r\\nfunction setFramework(value) {\\r\\n    framework = value;\\r\\n}\\r\\nfunction animate() {\\r\\n    translation = -translation;\\r\\n    let el = document.getElementById('hover-1');\\r\\n    el.style.transform = `translateY(${translation}px)`;\\r\\n    let el2 = document.getElementById('hover-2');\\r\\n    el2.style.transform = `translateY(${-translation}px)`;\\r\\n    setTimeout(animate, 9000);\\r\\n}\\r\\nlet translation = 25;\\r\\nonMount(() => {\\r\\n    setTimeout(animate, 10);\\r\\n});\\r\\nlet framework = \\\"Django\\\";\\r\\n</script>\\r\\n\\r\\n<svelte:head>\\r\\n  {@html codeStyle}\\r\\n</svelte:head>\\r\\n\\r\\n<div class=\\\"home\\\">\\r\\n  <div class=\\\"landing-page-container\\\">\\r\\n    <div class=\\\"landing-page\\\">\\r\\n      <div class=\\\"left\\\">\\r\\n        <h1>API Analytics</h1>\\r\\n        <h2>Monitoring and analytics for API frameworks.</h2>\\r\\n        <div class=\\\"links\\\">\\r\\n          <a href=\\\"/generate\\\" class=\\\"link\\\">\\r\\n            <div class=\\\"text\\\">\\r\\n              Try now - it's <span class=\\\"italic\\\">free</span>\\r\\n            </div></a\\r\\n          >\\r\\n          <a href=\\\"/dashboard/demo\\\" class=\\\"link secondary\\\">\\r\\n            <div class=\\\"text\\\">Demo</div></a\\r\\n          >\\r\\n        </div>\\r\\n      </div>\\r\\n      <div style=\\\"position: relative\\\" class=\\\"right\\\">\\r\\n        <img class=\\\"logo\\\" src=\\\"img/home-logo2.png\\\" alt=\\\"\\\" />\\r\\n        <img id=\\\"hover-1\\\" style=\\\"position: absolute;\\\" class=\\\"logo\\\" src=\\\"img/animated5.png\\\" alt=\\\"\\\" />\\r\\n        <img id=\\\"hover-2\\\" style=\\\"position: absolute;\\\" class=\\\"logo\\\" src=\\\"img/animated6.png\\\" alt=\\\"\\\" />\\r\\n      </div>\\r\\n    </div>\\r\\n  </div>\\r\\n\\r\\n  <div class=\\\"dashboard\\\">\\r\\n    <div class=\\\"dashboard-title-container\\\">\\r\\n      <img class=\\\"lightning-top\\\" src=\\\"img/logo.png\\\" alt=\\\"\\\" />\\r\\n      <h1 class=\\\"dashboard-title\\\">Dashboard</h1>\\r\\n    </div>\\r\\n    <div class=\\\"dashboard-content\\\">\\r\\n      <div class=\\\"dashboard-content-text\\\">\\r\\n        An all-in-one analytics dashboard. Real-time insight into your API's\\r\\n        usage.\\r\\n      </div>\\r\\n      <div class=\\\"dashboard-btn-container\\\">\\r\\n        <a href=\\\"/dashboard\\\" class=\\\"dashboard-btn secondary\\\">\\r\\n          <div class=\\\"dashboard-btn-text\\\">Open</div>\\r\\n        </a>\\r\\n      </div>\\r\\n    </div>\\r\\n    <img class=\\\"dashboard-img\\\" src=\\\"img/dashboard.png\\\" alt=\\\"\\\" />\\r\\n  </div>\\r\\n  <div class=\\\"dashboard\\\">\\r\\n    <div class=\\\"dashboard-title-container\\\">\\r\\n      <img class=\\\"lightning-top\\\" src=\\\"img/logo.png\\\" alt=\\\"\\\" />\\r\\n      <h1 class=\\\"dashboard-title\\\">Monitoring</h1>\\r\\n    </div>\\r\\n    <div class=\\\"dashboard-content\\\">\\r\\n      <div class=\\\"dashboard-content-text\\\">\\r\\n        Active monitoring and error notifications. Peace of mind.\\r\\n      </div>\\r\\n      <div class=\\\"dashboard-btn-container\\\">\\r\\n        <a href=\\\"/\\\" class=\\\"dashboard-btn secondary\\\">\\r\\n          <div class=\\\"dashboard-btn-text\\\">Coming Soon</div>\\r\\n        </a>\\r\\n      </div>\\r\\n    </div>\\r\\n    <img class=\\\"dashboard-img\\\" src=\\\"img/monitoring.png\\\" alt=\\\"\\\" />\\r\\n  </div>\\r\\n  <div class=\\\"add-middleware\\\">\\r\\n    <div class=\\\"add-middleware-title\\\">Getting Started</div>\\r\\n    <div class=\\\"frameworks\\\">\\r\\n      <button\\r\\n        class=\\\"framework python\\\"\\r\\n        class:active={framework == \\\"Django\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Django\\\");\\r\\n        }}>Django</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework python\\\"\\r\\n        class:active={framework == \\\"Flask\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Flask\\\");\\r\\n        }}>Flask</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework python\\\"\\r\\n        class:active={framework == \\\"FastAPI\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"FastAPI\\\");\\r\\n        }}>FastAPI</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework python\\\"\\r\\n        class:active={framework == \\\"Tornado\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Tornado\\\");\\r\\n        }}>Tornado</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework javascript\\\"\\r\\n        class:active={framework == \\\"Express\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Express\\\");\\r\\n        }}>Express</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework javascript\\\"\\r\\n        class:active={framework == \\\"Fastify\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Fastify\\\");\\r\\n        }}>Fastify</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework javascript\\\"\\r\\n        class:active={framework == \\\"Koa\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Koa\\\");\\r\\n        }}>Koa</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework golang\\\"\\r\\n        class:active={framework == \\\"Gin\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Gin\\\");\\r\\n        }}>Gin</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework golang\\\"\\r\\n        class:active={framework == \\\"Echo\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Echo\\\");\\r\\n        }}>Echo</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework golang\\\"\\r\\n        class:active={framework == \\\"Fiber\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Fiber\\\");\\r\\n        }}>Fiber</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework golang\\\"\\r\\n        class:active={framework == \\\"Chi\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Chi\\\");\\r\\n        }}>Chi</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework rust\\\"\\r\\n        class:active={framework == \\\"Actix\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Actix\\\");\\r\\n        }}>Actix</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework rust\\\"\\r\\n        class:active={framework == \\\"Axum\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Axum\\\");\\r\\n        }}>Axum</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework ruby\\\"\\r\\n        class:active={framework == \\\"Rails\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Rails\\\");\\r\\n        }}>Rails</button\\r\\n      >\\r\\n      <button\\r\\n        class=\\\"framework ruby\\\"\\r\\n        class:active={framework == \\\"Sinatra\\\"}\\r\\n        on:click={() => {\\r\\n          setFramework(\\\"Sinatra\\\");\\r\\n        }}>Sinatra</button\\r\\n      >\\r\\n    </div>\\r\\n    <div class=\\\"add-middleware-content\\\">\\r\\n      <div class=\\\"instructions-container\\\">\\r\\n        <div class=\\\"instructions\\\">\\r\\n          <div class=\\\"subtitle\\\">Install</div>\\r\\n          <code class=\\\"installation\\\"\\r\\n            >{frameworkExamples[framework].install}</code\\r\\n          >\\r\\n          <div class=\\\"subtitle\\\">Add middleware to API</div>\\r\\n          <!-- Render all code snippets to apply one-time syntax highlighting -->\\r\\n          <!-- TODO: dynamic syntax highlight rendering to only render the \\r\\n            frameworks clicked on and reduce this code to one line -->\\r\\n          <div class=\\\"code-file\\\">{frameworkExamples[framework].codeFile}</div>\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-python\\\"\\r\\n            style=\\\"{framework == 'Django' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Django\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-python\\\"\\r\\n            style=\\\"{framework == 'FastAPI' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"FastAPI\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-python\\\"\\r\\n            style=\\\"{framework == 'Flask' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Flask\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-python\\\"\\r\\n            style=\\\"{framework == 'Tornado' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Tornado\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-javascript\\\"\\r\\n            style=\\\"{framework == 'Express' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Express\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-javascript\\\"\\r\\n            style=\\\"{framework == 'Fastify' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Fastify\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-javascript\\\"\\r\\n            style=\\\"{framework == 'Koa' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Koa\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-go\\\"\\r\\n            style=\\\"{framework == 'Gin' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Gin\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-go\\\"\\r\\n            style=\\\"{framework == 'Echo' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Echo\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-go\\\"\\r\\n            style=\\\"{framework == 'Fiber' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Fiber\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-go\\\"\\r\\n            style=\\\"{framework == 'Chi' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Chi\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-rust\\\"\\r\\n            style=\\\"{framework == 'Actix' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Actix\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-rust\\\"\\r\\n            style=\\\"{framework == 'Axum' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Axum\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-ruby\\\"\\r\\n            style=\\\"{framework == 'Rails' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Rails\\\"].example}</code\\r\\n          >\\r\\n          <code\\r\\n            id=\\\"code\\\"\\r\\n            class=\\\"code language-ruby\\\"\\r\\n            style=\\\"{framework == 'Sinatra' ? 'display: initial' : ''} \\\"\\r\\n            >{frameworkExamples[\\\"Sinatra\\\"].example}</code\\r\\n          >\\r\\n        </div>\\r\\n      </div>\\r\\n    </div>\\r\\n  </div>\\r\\n</div>\\r\\n<Footer />\\r\\n\\r\\n<style scoped>\\r\\n  .landing-page {\\r\\n    display: flex;\\r\\n    place-items: center;\\r\\n    padding-bottom: 6em;\\r\\n  }\\r\\n  .landing-page-container {\\r\\n    display: grid;\\r\\n    margin: 0 12% 0 15%;\\r\\n    min-height: 100vh;\\r\\n  }\\r\\n  .left {\\r\\n    flex: 1;\\r\\n    text-align: left;\\r\\n  }\\r\\n  .right {\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n  }\\r\\n  .logo {\\r\\n    max-width: 1400px;\\r\\n    width: 700px;\\r\\n    margin-bottom: -50px;\\r\\n  }\\r\\n\\r\\n  h1 {\\r\\n    font-size: 3.4em;\\r\\n  }\\r\\n\\r\\n  h2 {\\r\\n    color: white;\\r\\n  }\\r\\n\\r\\n  .links {\\r\\n    color: #707070;\\r\\n    display: flex;\\r\\n    margin-top: 30px;\\r\\n    text-align: left;\\r\\n  }\\r\\n  .link {\\r\\n    width: fit-content;\\r\\n    margin-right: 20px;\\r\\n    font-size: 0.9;\\r\\n  }\\r\\n  .link:hover {\\r\\n    background: #31aa73;\\r\\n  }\\r\\n\\r\\n  .secondary {\\r\\n    background: #1c1c1c;\\r\\n    border: 3px solid var(--highlight);\\r\\n    color: var(--highlight);\\r\\n  }\\r\\n  .secondary:hover {\\r\\n    background: #081d13;\\r\\n  }\\r\\n\\r\\n  .lightning-top {\\r\\n    height: 70px;\\r\\n  }\\r\\n\\r\\n  .dashboard-title-container {\\r\\n    display: flex;\\r\\n    margin: 2em 4em;\\r\\n  }\\r\\n\\r\\n  a {\\r\\n    background: var(--highlight);\\r\\n    color: black;\\r\\n    padding: 10px 20px;\\r\\n    border-radius: 4px;\\r\\n  }\\r\\n  .italic {\\r\\n    font-style: italic;\\r\\n  }\\r\\n  .dashboard {\\r\\n    border: 3px solid var(--highlight);\\r\\n    width: 80%;\\r\\n    border-radius: 10px;\\r\\n    margin: auto;\\r\\n    margin-bottom: 8em;\\r\\n    overflow: hidden;\\r\\n    position: relative;\\r\\n  }\\r\\n  .dashboard-title {\\r\\n    font-size: 2.5em;\\r\\n    text-align: left;\\r\\n    margin: 0.2em 1em auto 1em;\\r\\n  }\\r\\n  .dashboard-title-container {\\r\\n    place-content: center;\\r\\n  }\\r\\n  .dashboard-img {\\r\\n    width: 81%;\\r\\n    border-radius: 10px;\\r\\n    box-shadow: 0px 24px 120px -25px var(--highlight);\\r\\n    margin-bottom: -1%;\\r\\n  }\\r\\n  .dashboard-content {\\r\\n    text-align: left;\\r\\n    color: white;\\r\\n    margin: 0 5em 4em;\\r\\n  }\\r\\n  .dashboard-content-text {\\r\\n    font-size: 1.1em;\\r\\n    text-align: center;\\r\\n  }\\r\\n  .dashboard-btn-container {\\r\\n    display: flex;\\r\\n    justify-content: center;\\r\\n    margin-top: 2em;\\r\\n  }\\r\\n  .dashboard-btn-text {\\r\\n    text-align: center;\\r\\n  }\\r\\n\\r\\n  .add-middleware-title {\\r\\n    color: var(--highlight);\\r\\n    font-size: 2.1em;\\r\\n    font-weight: 700;\\r\\n    margin-bottom: 1.5em;\\r\\n  }\\r\\n  .add-middleware {\\r\\n    margin: auto;\\r\\n    margin-bottom: 7em;\\r\\n    border-radius: 6px;\\r\\n  }\\r\\n\\r\\n  .frameworks {\\r\\n    margin: 0 10%;\\r\\n    overflow-x: auto;\\r\\n  }\\r\\n  .framework {\\r\\n    color: grey;\\r\\n    background: transparent;\\r\\n    border: none;\\r\\n    font-size: 1em;\\r\\n    cursor: pointer;\\r\\n    padding: 10px 18px;\\r\\n    border-bottom: 3px solid transparent;\\r\\n  }\\r\\n  .active {\\r\\n    color: white;\\r\\n  }\\r\\n  .active.python {\\r\\n    border-bottom: 3px solid #4b8bbe;\\r\\n  }\\r\\n  .active.golang {\\r\\n    border-bottom: 3px solid #00a7d0;\\r\\n  }\\r\\n  .active.javascript {\\r\\n    border-bottom: 3px solid #edd718;\\r\\n  }\\r\\n  .active.rust {\\r\\n    border-bottom: 3px solid #ef4900;\\r\\n  }\\r\\n  .active.ruby {\\r\\n    border-bottom: 3px solid #cd0000;\\r\\n  }\\r\\n  .subtitle {\\r\\n    color: rgb(110, 110, 110);\\r\\n    margin: 10px 0 2px 20px;\\r\\n    font-size: 0.85em;\\r\\n  }\\r\\n  .instructions-container {\\r\\n    padding: 1.5em 2em 2em;\\r\\n    width: 850px;\\r\\n    margin: auto;\\r\\n  }\\r\\n  .instructions {\\r\\n    text-align: left;\\r\\n    display: flex;\\r\\n    flex-direction: column;\\r\\n    position: relative;\\r\\n  }\\r\\n  code {\\r\\n    background: #151515;\\r\\n    padding: 1.4em 2em;\\r\\n    border-radius: 0.5em;\\r\\n    margin: 5px;\\r\\n    color: #dcdfe4;\\r\\n    white-space: pre-wrap;\\r\\n  }\\r\\n  .code {\\r\\n    display: none;\\r\\n  }\\r\\n  .code-file {\\r\\n    position: absolute;\\r\\n    font-size: 0.8em;\\r\\n    top: 160px;\\r\\n    color: rgb(97, 97, 97);\\r\\n    text-align: right;\\r\\n    right: 2.5em;\\r\\n    margin-bottom: -2em;\\r\\n  }\\r\\n\\r\\n  #hover-1 {\\r\\n    transform: translateY(30px);\\r\\n\\r\\n  }\\r\\n  #hover-2 {\\r\\n    transform: translateY(-30px);\\r\\n  }\\r\\n\\r\\n  img {\\r\\n   transition:all 10s ease-in-out;\\r\\n  }\\r\\n\\r\\n  @media screen and (max-width: 1500px) {\\r\\n    .landing-page-container {\\r\\n      margin: 0 6% 0 7%;\\r\\n    }\\r\\n  }\\r\\n  @media screen and (max-width: 1300px) {\\r\\n    .landing-page-container {\\r\\n      margin: 0 5% 0 6%;\\r\\n    }\\r\\n  }\\r\\n  @media screen and (max-width: 1200px) {\\r\\n    .landing-page {\\r\\n      flex-direction: column-reverse;\\r\\n    }\\r\\n    .landing-page-container {\\r\\n      margin: 0 2em;\\r\\n    }\\r\\n    .dashboard {\\r\\n      width: 90%;\\r\\n      margin-bottom: 4em;\\r\\n    }\\r\\n    .logo {\\r\\n      width: 100%;\\r\\n    }\\r\\n  }\\r\\n\\r\\n  @media screen and (max-width: 900px) {\\r\\n    .home {\\r\\n      font-size: 0.85em;\\r\\n    }\\r\\n    .instructions-container {\\r\\n      width: auto;\\r\\n    }\\r\\n    .code-file {\\r\\n      top: 140px;\\r\\n    }\\r\\n  }\\r\\n  @media screen and (max-width: 800px) {\\r\\n    .home {\\r\\n      font-size: 0.8em;\\r\\n    }\\r\\n    .right {\\r\\n      margin-top: 2em;\\r\\n    }\\r\\n  }\\r\\n  @media screen and (max-width: 700px) {\\r\\n    h1 {\\r\\n      font-size: 2.5em;\\r\\n    }\\r\\n    h2 {\\r\\n      font-size: 1.2em;\\r\\n    }\\r\\n    .landing-page-container {\\r\\n      min-height: unset;\\r\\n    }\\r\\n    .landing-page {\\r\\n      padding-bottom: 8em;\\r\\n    }\\r\\n    .lightning-top {\\r\\n      height: 55px;\\r\\n    }\\r\\n    .dashboard-img {\\r\\n      margin-bottom: -3%;\\r\\n    }\\r\\n    .add-middleware-title {\\r\\n      margin-bottom: 0.5em;\\r\\n    }\\r\\n    .instructions-container {\\r\\n      padding-top: 0;\\r\\n    }\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAiTE,aAAa,cAAC,CAAC,AACb,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,CACnB,cAAc,CAAE,GAAG,AACrB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,OAAO,CAAE,IAAI,CACb,MAAM,CAAE,CAAC,CAAC,GAAG,CAAC,CAAC,CAAC,GAAG,CACnB,UAAU,CAAE,KAAK,AACnB,CAAC,AACD,KAAK,cAAC,CAAC,AACL,IAAI,CAAE,CAAC,CACP,UAAU,CAAE,IAAI,AAClB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,KAAK,cAAC,CAAC,AACL,SAAS,CAAE,MAAM,CACjB,KAAK,CAAE,KAAK,CACZ,aAAa,CAAE,KAAK,AACtB,CAAC,AAED,EAAE,cAAC,CAAC,AACF,SAAS,CAAE,KAAK,AAClB,CAAC,AAED,EAAE,cAAC,CAAC,AACF,KAAK,CAAE,KAAK,AACd,CAAC,AAED,MAAM,cAAC,CAAC,AACN,KAAK,CAAE,OAAO,CACd,OAAO,CAAE,IAAI,CACb,UAAU,CAAE,IAAI,CAChB,UAAU,CAAE,IAAI,AAClB,CAAC,AACD,KAAK,cAAC,CAAC,AACL,KAAK,CAAE,WAAW,CAClB,YAAY,CAAE,IAAI,CAClB,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,mBAAK,MAAM,AAAC,CAAC,AACX,UAAU,CAAE,OAAO,AACrB,CAAC,AAED,UAAU,cAAC,CAAC,AACV,UAAU,CAAE,OAAO,CACnB,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,IAAI,WAAW,CAAC,CAClC,KAAK,CAAE,IAAI,WAAW,CAAC,AACzB,CAAC,AACD,wBAAU,MAAM,AAAC,CAAC,AAChB,UAAU,CAAE,OAAO,AACrB,CAAC,AAED,cAAc,cAAC,CAAC,AACd,MAAM,CAAE,IAAI,AACd,CAAC,AAED,0BAA0B,cAAC,CAAC,AAC1B,OAAO,CAAE,IAAI,CACb,MAAM,CAAE,GAAG,CAAC,GAAG,AACjB,CAAC,AAED,CAAC,cAAC,CAAC,AACD,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,KAAK,CAAE,KAAK,CACZ,OAAO,CAAE,IAAI,CAAC,IAAI,CAClB,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,OAAO,cAAC,CAAC,AACP,UAAU,CAAE,MAAM,AACpB,CAAC,AACD,UAAU,cAAC,CAAC,AACV,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,IAAI,WAAW,CAAC,CAClC,KAAK,CAAE,GAAG,CACV,aAAa,CAAE,IAAI,CACnB,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,QAAQ,CAAE,MAAM,CAChB,QAAQ,CAAE,QAAQ,AACpB,CAAC,AACD,gBAAgB,cAAC,CAAC,AAChB,SAAS,CAAE,KAAK,CAChB,UAAU,CAAE,IAAI,CAChB,MAAM,CAAE,KAAK,CAAC,GAAG,CAAC,IAAI,CAAC,GAAG,AAC5B,CAAC,AACD,0BAA0B,cAAC,CAAC,AAC1B,aAAa,CAAE,MAAM,AACvB,CAAC,AACD,cAAc,cAAC,CAAC,AACd,KAAK,CAAE,GAAG,CACV,aAAa,CAAE,IAAI,CACnB,UAAU,CAAE,GAAG,CAAC,IAAI,CAAC,KAAK,CAAC,KAAK,CAAC,IAAI,WAAW,CAAC,CACjD,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,kBAAkB,cAAC,CAAC,AAClB,UAAU,CAAE,IAAI,CAChB,KAAK,CAAE,KAAK,CACZ,MAAM,CAAE,CAAC,CAAC,GAAG,CAAC,GAAG,AACnB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,SAAS,CAAE,KAAK,CAChB,UAAU,CAAE,MAAM,AACpB,CAAC,AACD,wBAAwB,cAAC,CAAC,AACxB,OAAO,CAAE,IAAI,CACb,eAAe,CAAE,MAAM,CACvB,UAAU,CAAE,GAAG,AACjB,CAAC,AACD,mBAAmB,cAAC,CAAC,AACnB,UAAU,CAAE,MAAM,AACpB,CAAC,AAED,qBAAqB,cAAC,CAAC,AACrB,KAAK,CAAE,IAAI,WAAW,CAAC,CACvB,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,CAChB,aAAa,CAAE,KAAK,AACtB,CAAC,AACD,eAAe,cAAC,CAAC,AACf,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,aAAa,CAAE,GAAG,AACpB,CAAC,AAED,WAAW,cAAC,CAAC,AACX,MAAM,CAAE,CAAC,CAAC,GAAG,CACb,UAAU,CAAE,IAAI,AAClB,CAAC,AACD,UAAU,cAAC,CAAC,AACV,KAAK,CAAE,IAAI,CACX,UAAU,CAAE,WAAW,CACvB,MAAM,CAAE,IAAI,CACZ,SAAS,CAAE,GAAG,CACd,MAAM,CAAE,OAAO,CACf,OAAO,CAAE,IAAI,CAAC,IAAI,CAClB,aAAa,CAAE,GAAG,CAAC,KAAK,CAAC,WAAW,AACtC,CAAC,AACD,OAAO,cAAC,CAAC,AACP,KAAK,CAAE,KAAK,AACd,CAAC,AACD,OAAO,OAAO,cAAC,CAAC,AACd,aAAa,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,AAClC,CAAC,AACD,OAAO,OAAO,cAAC,CAAC,AACd,aAAa,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,AAClC,CAAC,AACD,OAAO,WAAW,cAAC,CAAC,AAClB,aAAa,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,AAClC,CAAC,AACD,OAAO,KAAK,cAAC,CAAC,AACZ,aAAa,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,AAClC,CAAC,AACD,OAAO,KAAK,cAAC,CAAC,AACZ,aAAa,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,AAClC,CAAC,AACD,SAAS,cAAC,CAAC,AACT,KAAK,CAAE,IAAI,GAAG,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAAC,CACzB,MAAM,CAAE,IAAI,CAAC,CAAC,CAAC,GAAG,CAAC,IAAI,CACvB,SAAS,CAAE,MAAM,AACnB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,OAAO,CAAE,KAAK,CAAC,GAAG,CAAC,GAAG,CACtB,KAAK,CAAE,KAAK,CACZ,MAAM,CAAE,IAAI,AACd,CAAC,AACD,aAAa,cAAC,CAAC,AACb,UAAU,CAAE,IAAI,CAChB,OAAO,CAAE,IAAI,CACb,cAAc,CAAE,MAAM,CACtB,QAAQ,CAAE,QAAQ,AACpB,CAAC,AACD,IAAI,cAAC,CAAC,AACJ,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,KAAK,CAAC,GAAG,CAClB,aAAa,CAAE,KAAK,CACpB,MAAM,CAAE,GAAG,CACX,KAAK,CAAE,OAAO,CACd,WAAW,CAAE,QAAQ,AACvB,CAAC,AACD,KAAK,cAAC,CAAC,AACL,OAAO,CAAE,IAAI,AACf,CAAC,AACD,UAAU,cAAC,CAAC,AACV,QAAQ,CAAE,QAAQ,CAClB,SAAS,CAAE,KAAK,CAChB,GAAG,CAAE,KAAK,CACV,KAAK,CAAE,IAAI,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,CACtB,UAAU,CAAE,KAAK,CACjB,KAAK,CAAE,KAAK,CACZ,aAAa,CAAE,IAAI,AACrB,CAAC,AAED,QAAQ,cAAC,CAAC,AACR,SAAS,CAAE,WAAW,IAAI,CAAC,AAE7B,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,SAAS,CAAE,WAAW,KAAK,CAAC,AAC9B,CAAC,AAED,GAAG,cAAC,CAAC,AACJ,WAAW,GAAG,CAAC,GAAG,CAAC,WAAW,AAC/B,CAAC,AAED,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,uBAAuB,cAAC,CAAC,AACvB,MAAM,CAAE,CAAC,CAAC,EAAE,CAAC,CAAC,CAAC,EAAE,AACnB,CAAC,AACH,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,uBAAuB,cAAC,CAAC,AACvB,MAAM,CAAE,CAAC,CAAC,EAAE,CAAC,CAAC,CAAC,EAAE,AACnB,CAAC,AACH,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,aAAa,cAAC,CAAC,AACb,cAAc,CAAE,cAAc,AAChC,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,MAAM,CAAE,CAAC,CAAC,GAAG,AACf,CAAC,AACD,UAAU,cAAC,CAAC,AACV,KAAK,CAAE,GAAG,CACV,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,KAAK,cAAC,CAAC,AACL,KAAK,CAAE,IAAI,AACb,CAAC,AACH,CAAC,AAED,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,KAAK,CAAC,AAAC,CAAC,AACpC,KAAK,cAAC,CAAC,AACL,SAAS,CAAE,MAAM,AACnB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,KAAK,CAAE,IAAI,AACb,CAAC,AACD,UAAU,cAAC,CAAC,AACV,GAAG,CAAE,KAAK,AACZ,CAAC,AACH,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,KAAK,CAAC,AAAC,CAAC,AACpC,KAAK,cAAC,CAAC,AACL,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,UAAU,CAAE,GAAG,AACjB,CAAC,AACH,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,KAAK,CAAC,AAAC,CAAC,AACpC,EAAE,cAAC,CAAC,AACF,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,EAAE,cAAC,CAAC,AACF,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,UAAU,CAAE,KAAK,AACnB,CAAC,AACD,aAAa,cAAC,CAAC,AACb,cAAc,CAAE,GAAG,AACrB,CAAC,AACD,cAAc,cAAC,CAAC,AACd,MAAM,CAAE,IAAI,AACd,CAAC,AACD,cAAc,cAAC,CAAC,AACd,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,qBAAqB,cAAC,CAAC,AACrB,aAAa,CAAE,KAAK,AACtB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,WAAW,CAAE,CAAC,AAChB,CAAC,AACH,CAAC\"}"
};

const Home = create_ssr_component(($$result, $$props, $$bindings, slots) => {

	function animate() {
		translation = -translation;
		let el = document.getElementById('hover-1');
		el.style.transform = `translateY(${translation}px)`;
		let el2 = document.getElementById('hover-2');
		el2.style.transform = `translateY(${-translation}px)`;
		setTimeout(animate, 9000);
	}

	let translation = 25;

	onMount(() => {
		setTimeout(animate, 10);
	});

	let framework = "Django";
	$$result.css.add(css$m);

	return `${($$result.head += `${a11yDark}`, "")}

<div class="${"home svelte-qzx5tq"}"><div class="${"landing-page-container svelte-qzx5tq"}"><div class="${"landing-page svelte-qzx5tq"}"><div class="${"left svelte-qzx5tq"}"><h1 class="${"svelte-qzx5tq"}">API Analytics</h1>
        <h2 class="${"svelte-qzx5tq"}">Monitoring and analytics for API frameworks.</h2>
        <div class="${"links svelte-qzx5tq"}"><a href="${"/generate"}" class="${"link svelte-qzx5tq"}"><div class="${"text"}">Try now - it&#39;s <span class="${"italic svelte-qzx5tq"}">free</span></div></a>
          <a href="${"/dashboard/demo"}" class="${"link secondary svelte-qzx5tq"}"><div class="${"text"}">Demo</div></a></div></div>
      <div style="${"position: relative"}" class="${"right svelte-qzx5tq"}"><img class="${"logo svelte-qzx5tq"}" src="${"img/home-logo2.png"}" alt="${""}">
        <img id="${"hover-1"}" style="${"position: absolute;"}" class="${"logo svelte-qzx5tq"}" src="${"img/animated5.png"}" alt="${""}">
        <img id="${"hover-2"}" style="${"position: absolute;"}" class="${"logo svelte-qzx5tq"}" src="${"img/animated6.png"}" alt="${""}"></div></div></div>

  <div class="${"dashboard svelte-qzx5tq"}"><div class="${"dashboard-title-container svelte-qzx5tq"}"><img class="${"lightning-top svelte-qzx5tq"}" src="${"img/logo.png"}" alt="${""}">
      <h1 class="${"dashboard-title svelte-qzx5tq"}">Dashboard</h1></div>
    <div class="${"dashboard-content svelte-qzx5tq"}"><div class="${"dashboard-content-text svelte-qzx5tq"}">An all-in-one analytics dashboard. Real-time insight into your API&#39;s
        usage.
      </div>
      <div class="${"dashboard-btn-container svelte-qzx5tq"}"><a href="${"/dashboard"}" class="${"dashboard-btn secondary svelte-qzx5tq"}"><div class="${"dashboard-btn-text svelte-qzx5tq"}">Open</div></a></div></div>
    <img class="${"dashboard-img svelte-qzx5tq"}" src="${"img/dashboard.png"}" alt="${""}"></div>
  <div class="${"dashboard svelte-qzx5tq"}"><div class="${"dashboard-title-container svelte-qzx5tq"}"><img class="${"lightning-top svelte-qzx5tq"}" src="${"img/logo.png"}" alt="${""}">
      <h1 class="${"dashboard-title svelte-qzx5tq"}">Monitoring</h1></div>
    <div class="${"dashboard-content svelte-qzx5tq"}"><div class="${"dashboard-content-text svelte-qzx5tq"}">Active monitoring and error notifications. Peace of mind.
      </div>
      <div class="${"dashboard-btn-container svelte-qzx5tq"}"><a href="${"/"}" class="${"dashboard-btn secondary svelte-qzx5tq"}"><div class="${"dashboard-btn-text svelte-qzx5tq"}">Coming Soon</div></a></div></div>
    <img class="${"dashboard-img svelte-qzx5tq"}" src="${"img/monitoring.png"}" alt="${""}"></div>
  <div class="${"add-middleware svelte-qzx5tq"}"><div class="${"add-middleware-title svelte-qzx5tq"}">Getting Started</div>
    <div class="${"frameworks svelte-qzx5tq"}"><button class="${["framework python svelte-qzx5tq", "active" ].join(' ').trim()}">Django</button>
      <button class="${["framework python svelte-qzx5tq", ""].join(' ').trim()}">Flask</button>
      <button class="${["framework python svelte-qzx5tq", ""].join(' ').trim()}">FastAPI</button>
      <button class="${["framework python svelte-qzx5tq", ""].join(' ').trim()}">Tornado</button>
      <button class="${["framework javascript svelte-qzx5tq", ""].join(' ').trim()}">Express</button>
      <button class="${["framework javascript svelte-qzx5tq", ""].join(' ').trim()}">Fastify</button>
      <button class="${["framework javascript svelte-qzx5tq", ""].join(' ').trim()}">Koa</button>
      <button class="${["framework golang svelte-qzx5tq", ""].join(' ').trim()}">Gin</button>
      <button class="${["framework golang svelte-qzx5tq", ""].join(' ').trim()}">Echo</button>
      <button class="${["framework golang svelte-qzx5tq", ""].join(' ').trim()}">Fiber</button>
      <button class="${["framework golang svelte-qzx5tq", ""].join(' ').trim()}">Chi</button>
      <button class="${["framework rust svelte-qzx5tq", ""].join(' ').trim()}">Actix</button>
      <button class="${["framework rust svelte-qzx5tq", ""].join(' ').trim()}">Axum</button>
      <button class="${["framework ruby svelte-qzx5tq", ""].join(' ').trim()}">Rails</button>
      <button class="${["framework ruby svelte-qzx5tq", ""].join(' ').trim()}">Sinatra</button></div>
    <div class="${"add-middleware-content"}"><div class="${"instructions-container svelte-qzx5tq"}"><div class="${"instructions svelte-qzx5tq"}"><div class="${"subtitle svelte-qzx5tq"}">Install</div>
          <code class="${"installation svelte-qzx5tq"}">${escape(frameworkExamples[framework].install)}</code>
          <div class="${"subtitle svelte-qzx5tq"}">Add middleware to API</div>
          
          
          <div class="${"code-file svelte-qzx5tq"}">${escape(frameworkExamples[framework].codeFile)}</div>
          <code id="${"code"}" class="${"code language-python svelte-qzx5tq"}" style="${escape('display: initial' , true) + ""}">${escape(frameworkExamples["Django"].example)}</code>
          <code id="${"code"}" class="${"code language-python svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["FastAPI"].example)}</code>
          <code id="${"code"}" class="${"code language-python svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Flask"].example)}</code>
          <code id="${"code"}" class="${"code language-python svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Tornado"].example)}</code>
          <code id="${"code"}" class="${"code language-javascript svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Express"].example)}</code>
          <code id="${"code"}" class="${"code language-javascript svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Fastify"].example)}</code>
          <code id="${"code"}" class="${"code language-javascript svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Koa"].example)}</code>
          <code id="${"code"}" class="${"code language-go svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Gin"].example)}</code>
          <code id="${"code"}" class="${"code language-go svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Echo"].example)}</code>
          <code id="${"code"}" class="${"code language-go svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Fiber"].example)}</code>
          <code id="${"code"}" class="${"code language-go svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Chi"].example)}</code>
          <code id="${"code"}" class="${"code language-rust svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Actix"].example)}</code>
          <code id="${"code"}" class="${"code language-rust svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Axum"].example)}</code>
          <code id="${"code"}" class="${"code language-ruby svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Rails"].example)}</code>
          <code id="${"code"}" class="${"code language-ruby svelte-qzx5tq"}" style="${escape('', true) + ""}">${escape(frameworkExamples["Sinatra"].example)}</code></div></div></div></div></div>
${validate_component(Footer, "Footer").$$render($$result, {}, {}, {})}`;
});

/* src\routes\Generate.svelte generated by Svelte v3.53.1 */

const css$l = {
	code: ".generate.svelte-r5bis2{display:grid;place-items:center}h2.svelte-r5bis2{margin:0 0 1em;font-size:2em}.content.svelte-r5bis2{width:fit-content;background:#343434;background:#1c1c1c;padding:3.5em 4em 4em;border-radius:9px;margin:20vh 0 2vh;height:400px}input.svelte-r5bis2{background:#1c1c1c;background:#343434;border:none;padding:0 20px;width:310px;font-size:1em;text-align:center;height:40px;border-radius:4px;margin-bottom:2.5em;color:white;display:grid}button.svelte-r5bis2{height:40px;border-radius:4px;padding:0 20px;border:none;cursor:pointer;width:100px}.highlight.svelte-r5bis2{color:#3fcf8e}.details.svelte-r5bis2{font-size:0.8em}.keep-secure.svelte-r5bis2{color:#5a5a5a;margin-bottom:1em}#copyBtn.svelte-r5bis2{background:#1c1c1c;display:none;background:#343434;place-items:center;margin:auto}.copy-icon.svelte-r5bis2{filter:contrast(0.3);height:20px}#copied.svelte-r5bis2{color:var(--highlight);margin:2em auto auto;visibility:hidden;height:1em}.spinner.svelte-r5bis2{height:7em}#generateBtn.svelte-r5bis2{background:#3fcf8e}",
	map: "{\"version\":3,\"file\":\"Generate.svelte\",\"sources\":[\"Generate.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">let loading = false;\\r\\nlet generatedKey = false;\\r\\nlet apiKey = \\\"\\\";\\r\\nlet generateBtn;\\r\\nlet copyBtn;\\r\\nlet copiedNotification;\\r\\nasync function genAPIKey() {\\r\\n    if (!generatedKey) {\\r\\n        loading = true;\\r\\n        const response = await fetch(\\\"https://api-analytics-server.vercel.app/api/generate-api-key\\\");\\r\\n        if (response.status == 200) {\\r\\n            const data = await response.json();\\r\\n            generatedKey = true;\\r\\n            apiKey = data.value;\\r\\n            generateBtn.style.display = \\\"none\\\";\\r\\n            copyBtn.style.display = \\\"grid\\\";\\r\\n        }\\r\\n        loading = false;\\r\\n    }\\r\\n}\\r\\nfunction copyToClipboard() {\\r\\n    navigator.clipboard.writeText(apiKey);\\r\\n    copyBtn.value = \\\"Copied\\\";\\r\\n    copiedNotification.style.visibility = \\\"visible\\\";\\r\\n}\\r\\n</script>\\r\\n\\r\\n<div class=\\\"generate\\\">\\r\\n  <div class=\\\"content\\\">\\r\\n    <h2>Generate API key</h2>\\r\\n    <input type=\\\"text\\\" readonly bind:value={apiKey} />\\r\\n    <button id=\\\"generateBtn\\\" on:click={genAPIKey} bind:this={generateBtn}\\r\\n      >Generate</button\\r\\n    >\\r\\n    <button id=\\\"copyBtn\\\" on:click={copyToClipboard} bind:this={copyBtn}\\r\\n      ><img class=\\\"copy-icon\\\" src=\\\"img/copy.png\\\" alt=\\\"\\\" /></button\\r\\n    >\\r\\n    <div id=\\\"copied\\\" bind:this={copiedNotification}>Copied!</div>\\r\\n\\r\\n    <div class=\\\"spinner\\\">\\r\\n      <div class=\\\"loader\\\" style=\\\"display: {loading ? 'initial' : 'none'}\\\" />\\r\\n    </div>\\r\\n  </div>\\r\\n  <div class=\\\"details\\\">\\r\\n    <div class=\\\"keep-secure\\\">Keep your API key safe and secure.</div>\\r\\n    <div class=\\\"highlight logo\\\">API Analytics</div>\\r\\n    <img class=\\\"footer-logo\\\" src=\\\"img/logo.png\\\" alt=\\\"\\\" />\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .generate {\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n  }\\r\\n  h2 {\\r\\n    margin: 0 0 1em;\\r\\n    font-size: 2em;\\r\\n  }\\r\\n  .content {\\r\\n    width: fit-content;\\r\\n    background: #343434;\\r\\n    background: #1c1c1c;\\r\\n    padding: 3.5em 4em 4em;\\r\\n    border-radius: 9px;\\r\\n    margin: 20vh 0 2vh;\\r\\n    height: 400px;\\r\\n  }\\r\\n  input {\\r\\n    background: #1c1c1c;\\r\\n    background: #343434;\\r\\n    border: none;\\r\\n    padding: 0 20px;\\r\\n    width: 310px;\\r\\n    font-size: 1em;\\r\\n    text-align: center;\\r\\n    height: 40px;\\r\\n    border-radius: 4px;\\r\\n    margin-bottom: 2.5em;\\r\\n    color: white;\\r\\n    display: grid;\\r\\n  }\\r\\n  button {\\r\\n    height: 40px;\\r\\n    border-radius: 4px;\\r\\n    padding: 0 20px;\\r\\n    border: none;\\r\\n    cursor: pointer;\\r\\n    width: 100px;\\r\\n  }\\r\\n  .highlight {\\r\\n    color: #3fcf8e;\\r\\n  }\\r\\n  .details {\\r\\n    font-size: 0.8em;\\r\\n  }\\r\\n  .keep-secure {\\r\\n    color: #5a5a5a;\\r\\n    margin-bottom: 1em;\\r\\n  }\\r\\n\\r\\n  #copyBtn {\\r\\n    background: #1c1c1c;\\r\\n    display: none;\\r\\n    background: #343434;\\r\\n    place-items: center;\\r\\n    margin: auto;\\r\\n  }\\r\\n  .copy-icon {\\r\\n    filter: contrast(0.3);\\r\\n    height: 20px;\\r\\n  }\\r\\n  #copied {\\r\\n    color: var(--highlight);\\r\\n    margin: 2em auto auto;\\r\\n    visibility: hidden;\\r\\n    height: 1em;\\r\\n  }\\r\\n\\r\\n  .spinner {\\r\\n    height: 7em;\\r\\n    /* margin-bottom: 5em; */\\r\\n  }\\r\\n  #generateBtn {\\r\\n    background: #3fcf8e;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAmDE,SAAS,cAAC,CAAC,AACT,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,EAAE,cAAC,CAAC,AACF,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CACf,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,KAAK,CAAE,WAAW,CAClB,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,KAAK,CAAC,GAAG,CAAC,GAAG,CACtB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,CAAC,CAAC,CAAC,GAAG,CAClB,MAAM,CAAE,KAAK,AACf,CAAC,AACD,KAAK,cAAC,CAAC,AACL,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,MAAM,CAAE,IAAI,CACZ,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,KAAK,CAAE,KAAK,CACZ,SAAS,CAAE,GAAG,CACd,UAAU,CAAE,MAAM,CAClB,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,aAAa,CAAE,KAAK,CACpB,KAAK,CAAE,KAAK,CACZ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,cAAC,CAAC,AACN,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,OAAO,CACf,KAAK,CAAE,KAAK,AACd,CAAC,AACD,UAAU,cAAC,CAAC,AACV,KAAK,CAAE,OAAO,AAChB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,YAAY,cAAC,CAAC,AACZ,KAAK,CAAE,OAAO,CACd,aAAa,CAAE,GAAG,AACpB,CAAC,AAED,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,IAAI,CACb,UAAU,CAAE,OAAO,CACnB,WAAW,CAAE,MAAM,CACnB,MAAM,CAAE,IAAI,AACd,CAAC,AACD,UAAU,cAAC,CAAC,AACV,MAAM,CAAE,SAAS,GAAG,CAAC,CACrB,MAAM,CAAE,IAAI,AACd,CAAC,AACD,OAAO,cAAC,CAAC,AACP,KAAK,CAAE,IAAI,WAAW,CAAC,CACvB,MAAM,CAAE,GAAG,CAAC,IAAI,CAAC,IAAI,CACrB,UAAU,CAAE,MAAM,CAClB,MAAM,CAAE,GAAG,AACb,CAAC,AAED,QAAQ,cAAC,CAAC,AACR,MAAM,CAAE,GAAG,AAEb,CAAC,AACD,YAAY,cAAC,CAAC,AACZ,UAAU,CAAE,OAAO,AACrB,CAAC\"}"
};

const Generate = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let apiKey = "";
	let generateBtn;
	let copyBtn;
	let copiedNotification;

	$$result.css.add(css$l);

	return `<div class="${"generate svelte-r5bis2"}"><div class="${"content svelte-r5bis2"}"><h2 class="${"svelte-r5bis2"}">Generate API key</h2>
    <input type="${"text"}" readonly class="${"svelte-r5bis2"}"${add_attribute("value", apiKey, 0)}>
    <button id="${"generateBtn"}" class="${"svelte-r5bis2"}"${add_attribute("this", generateBtn, 0)}>Generate</button>
    <button id="${"copyBtn"}" class="${"svelte-r5bis2"}"${add_attribute("this", copyBtn, 0)}><img class="${"copy-icon svelte-r5bis2"}" src="${"img/copy.png"}" alt="${""}"></button>
    <div id="${"copied"}" class="${"svelte-r5bis2"}"${add_attribute("this", copiedNotification, 0)}>Copied!</div>

    <div class="${"spinner svelte-r5bis2"}"><div class="${"loader"}" style="${"display: " + escape('none', true)}"></div></div></div>
  <div class="${"details svelte-r5bis2"}"><div class="${"keep-secure svelte-r5bis2"}">Keep your API key safe and secure.</div>
    <div class="${"highlight logo svelte-r5bis2"}">API Analytics</div>
    <img class="${"footer-logo"}" src="${"img/logo.png"}" alt="${""}"></div>
</div>`;
});

/* src\routes\SignIn.svelte generated by Svelte v3.53.1 */

const css$k = {
	code: ".generate.svelte-1kmr1cu{display:grid;place-items:center}h2.svelte-1kmr1cu{margin:0 0 1em;font-size:2em}.content.svelte-1kmr1cu{width:fit-content;background:#343434;background:#1c1c1c;padding:3.5em 4em 4em;border-radius:9px;margin:20vh 0 2vh;height:400px}input.svelte-1kmr1cu{background:#1c1c1c;background:#343434;border:none;padding:0 20px;width:310px;font-size:1em;text-align:center;height:40px;border-radius:4px;margin-bottom:2.5em;color:white;display:grid}button.svelte-1kmr1cu{height:40px;border-radius:4px;padding:0 20px;border:none;cursor:pointer;width:100px}.highlight.svelte-1kmr1cu{color:#3fcf8e}.details.svelte-1kmr1cu{font-size:0.8em}.keep-secure.svelte-1kmr1cu{color:#5a5a5a;margin-bottom:1em}#generateBtn.svelte-1kmr1cu{background:#3fcf8e}",
	map: "{\"version\":3,\"file\":\"SignIn.svelte\",\"sources\":[\"SignIn.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">let apiKey = \\\"\\\";\\r\\nlet loading = false;\\r\\nasync function genAPIKey() {\\r\\n    loading = true;\\r\\n    // Fetch page ID\\r\\n    const response = await fetch(`https://api-analytics-server.vercel.app/api/user-id/${apiKey}`);\\r\\n    console.log(response);\\r\\n    if (response.status == 200) {\\r\\n        const data = await response.json();\\r\\n        window.location.href = `/dashboard/${data.value.replaceAll(\\\"-\\\", \\\"\\\")}`;\\r\\n    }\\r\\n    loading = false;\\r\\n}\\r\\n</script>\\r\\n\\r\\n<div class=\\\"generate\\\">\\r\\n  <div class=\\\"content\\\">\\r\\n    <h2>Dashboard</h2>\\r\\n    <input type=\\\"text\\\" bind:value={apiKey} placeholder=\\\"Enter API key\\\"/>\\r\\n    <button id=\\\"generateBtn\\\" on:click={genAPIKey}>Load</button>\\r\\n    <div class=\\\"spinner\\\">\\r\\n      <div class=\\\"loader\\\" style=\\\"display: {loading ? 'initial' : 'none'}\\\" />\\r\\n    </div>\\r\\n  </div>\\r\\n  <div class=\\\"details\\\">\\r\\n    <div class=\\\"keep-secure\\\">Keep your API key safe and secure.</div>\\r\\n    <div class=\\\"highlight logo\\\">API Analytics</div>\\r\\n    <img class=\\\"footer-logo\\\" src=\\\"img/logo.png\\\" alt=\\\"\\\">\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .generate {\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n  }\\r\\n  h2 {\\r\\n    margin: 0 0 1em;\\r\\n    font-size: 2em;\\r\\n  }\\r\\n  .content {\\r\\n    width: fit-content;\\r\\n    background: #343434;\\r\\n    background: #1c1c1c;\\r\\n    padding: 3.5em 4em 4em;\\r\\n    border-radius: 9px;\\r\\n    margin: 20vh 0 2vh;\\r\\n    height: 400px;\\r\\n  }\\r\\n  input {\\r\\n    background: #1c1c1c;\\r\\n    background: #343434;\\r\\n    border: none;\\r\\n    padding: 0 20px;\\r\\n    width: 310px;\\r\\n    font-size: 1em;\\r\\n    text-align: center;\\r\\n    height: 40px;\\r\\n    border-radius: 4px;\\r\\n    margin-bottom: 2.5em;\\r\\n    color: white;\\r\\n    display: grid;\\r\\n  }\\r\\n  button {\\r\\n    height: 40px;\\r\\n    border-radius: 4px;\\r\\n    padding: 0 20px;\\r\\n    border: none;\\r\\n    cursor: pointer;\\r\\n    width: 100px;\\r\\n  }\\r\\n  .highlight {\\r\\n    color: #3fcf8e;\\r\\n  }\\r\\n  .details {\\r\\n    font-size: 0.8em;\\r\\n  }\\r\\n  .keep-secure {\\r\\n    color: #5a5a5a;\\r\\n    margin-bottom: 1em;\\r\\n  }\\r\\n  #generateBtn {\\r\\n    background: #3fcf8e;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAgCE,SAAS,eAAC,CAAC,AACT,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,EAAE,eAAC,CAAC,AACF,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CACf,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,KAAK,CAAE,WAAW,CAClB,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,KAAK,CAAC,GAAG,CAAC,GAAG,CACtB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,CAAC,CAAC,CAAC,GAAG,CAClB,MAAM,CAAE,KAAK,AACf,CAAC,AACD,KAAK,eAAC,CAAC,AACL,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,MAAM,CAAE,IAAI,CACZ,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,KAAK,CAAE,KAAK,CACZ,SAAS,CAAE,GAAG,CACd,UAAU,CAAE,MAAM,CAClB,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,aAAa,CAAE,KAAK,CACpB,KAAK,CAAE,KAAK,CACZ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,eAAC,CAAC,AACN,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,OAAO,CACf,KAAK,CAAE,KAAK,AACd,CAAC,AACD,UAAU,eAAC,CAAC,AACV,KAAK,CAAE,OAAO,AAChB,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,KAAK,CAAE,OAAO,CACd,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,UAAU,CAAE,OAAO,AACrB,CAAC\"}"
};

const SignIn = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let apiKey = "";

	$$result.css.add(css$k);

	return `<div class="${"generate svelte-1kmr1cu"}"><div class="${"content svelte-1kmr1cu"}"><h2 class="${"svelte-1kmr1cu"}">Dashboard</h2>
    <input type="${"text"}" placeholder="${"Enter API key"}" class="${"svelte-1kmr1cu"}"${add_attribute("value", apiKey, 0)}>
    <button id="${"generateBtn"}" class="${"svelte-1kmr1cu"}">Load</button>
    <div class="${"spinner"}"><div class="${"loader"}" style="${"display: " + escape('none', true)}"></div></div></div>
  <div class="${"details svelte-1kmr1cu"}"><div class="${"keep-secure svelte-1kmr1cu"}">Keep your API key safe and secure.</div>
    <div class="${"highlight logo svelte-1kmr1cu"}">API Analytics</div>
    <img class="${"footer-logo"}" src="${"img/logo.png"}" alt="${""}"></div>
</div>`;
});

/* src\components\dashboard\Requests.svelte generated by Svelte v3.53.1 */

const css$j = {
	code: ".card.svelte-1mgix01{width:calc(200px - 1em);margin:0 1em 0 2em;position:relative}.value.svelte-1mgix01{margin:20px 0;font-size:1.8em;font-weight:600}.percentage-change.svelte-1mgix01{position:absolute;right:20px;top:20px;font-size:0.8em}.positive.svelte-1mgix01{color:var(--highlight)}.negative.svelte-1mgix01{color:rgb(228, 97, 97)}",
	map: "{\"version\":3,\"file\":\"Requests.svelte\",\"sources\":[\"Requests.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction setPercentageChange() {\\r\\n    if (prevData.length == 0) {\\r\\n        percentageChange = null;\\r\\n    }\\r\\n    else {\\r\\n        percentageChange = (data.length / prevData.length) * 100 - 100;\\r\\n    }\\r\\n}\\r\\nlet percentageChange;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && setPercentageChange();\\r\\nexport let data, prevData;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Total\\\">\\r\\n  {#if percentageChange != null}\\r\\n    <div\\r\\n      class=\\\"percentage-change\\\"\\r\\n      class:positive={percentageChange > 0}\\r\\n      class:negative={percentageChange < 0}\\r\\n    >\\r\\n      ({percentageChange > 0 ? \\\"+\\\" : \\\"\\\"}{percentageChange.toFixed(1)}%)\\r\\n    </div>\\r\\n  {/if}\\r\\n  <div class=\\\"card-title\\\">Requests</div>\\r\\n  <div class=\\\"value\\\">{data.length.toLocaleString()}</div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    width: calc(200px - 1em);\\r\\n    margin: 0 1em 0 2em;\\r\\n    position: relative;\\r\\n  }\\r\\n  .value {\\r\\n    margin: 20px 0;\\r\\n    font-size: 1.8em;\\r\\n    font-weight: 600;\\r\\n  }\\r\\n  .percentage-change {\\r\\n    position: absolute;\\r\\n    right: 20px;\\r\\n    top: 20px;\\r\\n    font-size: 0.8em;\\r\\n  }\\r\\n  .positive {\\r\\n    color: var(--highlight);\\r\\n  }\\r\\n  .negative {\\r\\n    color: rgb(228, 97, 97);\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAiCE,KAAK,eAAC,CAAC,AACL,KAAK,CAAE,KAAK,KAAK,CAAC,CAAC,CAAC,GAAG,CAAC,CACxB,MAAM,CAAE,CAAC,CAAC,GAAG,CAAC,CAAC,CAAC,GAAG,CACnB,QAAQ,CAAE,QAAQ,AACpB,CAAC,AACD,MAAM,eAAC,CAAC,AACN,MAAM,CAAE,IAAI,CAAC,CAAC,CACd,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AACD,kBAAkB,eAAC,CAAC,AAClB,QAAQ,CAAE,QAAQ,CAClB,KAAK,CAAE,IAAI,CACX,GAAG,CAAE,IAAI,CACT,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,SAAS,eAAC,CAAC,AACT,KAAK,CAAE,IAAI,WAAW,CAAC,AACzB,CAAC,AACD,SAAS,eAAC,CAAC,AACT,KAAK,CAAE,IAAI,GAAG,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,AACzB,CAAC\"}"
};

const Requests = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function setPercentageChange() {
		if (prevData.length == 0) {
			percentageChange = null;
		} else {
			percentageChange = data.length / prevData.length * 100 - 100;
		}
	}

	let percentageChange;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data, prevData } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.prevData === void 0 && $$bindings.prevData && prevData !== void 0) $$bindings.prevData(prevData);
	$$result.css.add(css$j);
	data && mounted && setPercentageChange();

	return `<div class="${"card svelte-1mgix01"}" title="${"Total"}">${percentageChange != null
	? `<div class="${[
			"percentage-change svelte-1mgix01",
			(percentageChange > 0 ? "positive" : "") + ' ' + (percentageChange < 0 ? "negative" : "")
		].join(' ').trim()}">(${escape(percentageChange > 0 ? "+" : "")}${escape(percentageChange.toFixed(1))}%)
    </div>`
	: ``}
  <div class="${"card-title"}">Requests</div>
  <div class="${"value svelte-1mgix01"}">${escape(data.length.toLocaleString())}</div>
</div>`;
});

/* src\components\dashboard\Welcome.svelte generated by Svelte v3.53.1 */

const css$i = {
	code: ".card.svelte-1w2ck9z{width:calc(200px - 1em);margin:0 1em 2em 2em;background:transparent;display:grid;place-items:center}img.svelte-1w2ck9z{width:25px}",
	map: "{\"version\":3,\"file\":\"Welcome.svelte\",\"sources\":[\"Welcome.svelte\"],\"sourcesContent\":[\"<div class=\\\"card\\\" title=\\\"API Analytics\\\">\\r\\n  <img src=\\\"../img/logo.png\\\" alt=\\\"\\\" />\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    width: calc(200px - 1em);\\r\\n    margin: 0 1em 2em 2em;\\r\\n    background: transparent;\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n  }\\r\\n  img {\\r\\n    width: 25px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAKE,KAAK,eAAC,CAAC,AACL,KAAK,CAAE,KAAK,KAAK,CAAC,CAAC,CAAC,GAAG,CAAC,CACxB,MAAM,CAAE,CAAC,CAAC,GAAG,CAAC,GAAG,CAAC,GAAG,CACrB,UAAU,CAAE,WAAW,CACvB,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,GAAG,eAAC,CAAC,AACH,KAAK,CAAE,IAAI,AACb,CAAC\"}"
};

const Welcome = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	$$result.css.add(css$i);

	return `<div class="${"card svelte-1w2ck9z"}" title="${"API Analytics"}"><img src="${"../img/logo.png"}" alt="${""}" class="${"svelte-1w2ck9z"}">
</div>`;
});

function periodToDays$1(period) {
    if (period == "24-hours") {
        return 1;
    }
    else if (period == "week") {
        return 8;
    }
    else if (period == "month") {
        return 30;
    }
    else if (period == "3-months") {
        return 30 * 3;
    }
    else if (period == "6-months") {
        return 30 * 6;
    }
    else if (period == "year") {
        return 365;
    }
    else {
        return null;
    }
}

/* src\components\dashboard\RequestsPerHour.svelte generated by Svelte v3.53.1 */

const css$h = {
	code: ".card.svelte-r5qgcj{width:calc(200px - 1em);margin:0 2em 0 1em}.value.svelte-r5qgcj{margin:20px 0;font-size:1.8em;font-weight:600}.per-hour.svelte-r5qgcj{color:var(--dim-text);font-size:0.8em;margin-left:4px}",
	map: "{\"version\":3,\"file\":\"RequestsPerHour.svelte\",\"sources\":[\"RequestsPerHour.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nimport periodToDays from \\\"../../lib/period\\\";\\r\\nfunction build() {\\r\\n    let totalRequests = 0;\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        totalRequests++;\\r\\n    }\\r\\n    if (totalRequests > 0) {\\r\\n        let days = periodToDays(period);\\r\\n        if (days != null) {\\r\\n            requestsPerHour = (totalRequests / (24 * days)).toFixed(2);\\r\\n        }\\r\\n    }\\r\\n    else {\\r\\n        requestsPerHour = \\\"0\\\";\\r\\n    }\\r\\n}\\r\\nlet mounted = false;\\r\\nlet requestsPerHour;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && build();\\r\\nexport let data, period;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Last week\\\">\\r\\n  <div class=\\\"card-title\\\">\\r\\n    Requests <span class=\\\"per-hour\\\">/ hour</span>\\r\\n  </div>\\r\\n  {#if requestsPerHour != undefined}\\r\\n    <div class=\\\"value\\\">{requestsPerHour}</div>\\r\\n  {/if}\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    width: calc(200px - 1em);\\r\\n    margin: 0 2em 0 1em;\\r\\n  }\\r\\n  .value {\\r\\n    margin: 20px 0;\\r\\n    font-size: 1.8em;\\r\\n    font-weight: 600;\\r\\n  }\\r\\n  .per-hour {\\r\\n    color: var(--dim-text);\\r\\n    font-size: 0.8em;\\r\\n    margin-left: 4px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAoCE,KAAK,cAAC,CAAC,AACL,KAAK,CAAE,KAAK,KAAK,CAAC,CAAC,CAAC,GAAG,CAAC,CACxB,MAAM,CAAE,CAAC,CAAC,GAAG,CAAC,CAAC,CAAC,GAAG,AACrB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,MAAM,CAAE,IAAI,CAAC,CAAC,CACd,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AACD,SAAS,cAAC,CAAC,AACT,KAAK,CAAE,IAAI,UAAU,CAAC,CACtB,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC\"}"
};

const RequestsPerHour = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function build() {
		let totalRequests = 0;

		for (let i = 0; i < data.length; i++) {
			totalRequests++;
		}

		if (totalRequests > 0) {
			let days = periodToDays$1(period);

			if (days != null) {
				requestsPerHour = (totalRequests / (24 * days)).toFixed(2);
			}
		} else {
			requestsPerHour = "0";
		}
	}

	let mounted = false;
	let requestsPerHour;

	onMount(() => {
		mounted = true;
	});

	let { data, period } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.period === void 0 && $$bindings.period && period !== void 0) $$bindings.period(period);
	$$result.css.add(css$h);
	data && mounted && build();

	return `<div class="${"card svelte-r5qgcj"}" title="${"Last week"}"><div class="${"card-title"}">Requests <span class="${"per-hour svelte-r5qgcj"}">/ hour</span></div>
  ${requestsPerHour != undefined
	? `<div class="${"value svelte-r5qgcj"}">${escape(requestsPerHour)}</div>`
	: ``}
</div>`;
});

/* src\components\dashboard\ResponseTimes.svelte generated by Svelte v3.53.1 */

const css$g = {
	code: ".values.svelte-kx5j01{display:flex;color:var(--highlight);font-size:1.8em;font-weight:700}.values.svelte-kx5j01,.labels.svelte-kx5j01{margin:0 0.5rem}.value.svelte-kx5j01{flex:1;font-size:1.1em;padding:20px 20px 4px}.labels.svelte-kx5j01{display:flex;font-size:0.8em;color:var(--dim-text)}.label.svelte-kx5j01{flex:1}.milliseconds.svelte-kx5j01{color:var(--dim-text);font-size:0.8em;margin-left:4px}.median.svelte-kx5j01{font-size:1em}.upper-quartile.svelte-kx5j01,.lower-quartile.svelte-kx5j01{font-size:1em;padding-bottom:0}.bar.svelte-kx5j01{padding:20px 0 20px;display:flex;height:30px;width:85%;margin:auto;align-items:center;position:relative}.bar-green.svelte-kx5j01{background:var(--highlight);width:75%;height:10px;border-radius:2px 0 0 2px}.bar-yellow.svelte-kx5j01{width:15%;height:10px;background:rgb(235, 235, 129)}.bar-red.svelte-kx5j01{width:20%;height:10px;border-radius:0 2px 2px 0;background:rgb(228, 97, 97)}.marker.svelte-kx5j01{position:absolute;height:30px;width:5px;background:white;border-radius:2px;left:0}",
	map: "{\"version\":3,\"file\":\"ResponseTimes.svelte\",\"sources\":[\"ResponseTimes.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\n// Median and quartiles from StackOverflow answer\\r\\n// https://stackoverflow.com/a/55297611/8851732\\r\\nconst asc = (arr) => arr.sort((a, b) => a - b);\\r\\nconst sum = (arr) => arr.reduce((a, b) => a + b, 0);\\r\\nconst mean = (arr) => sum(arr) / arr.length;\\r\\n// sample standard deviation\\r\\nfunction std(arr) {\\r\\n    const mu = mean(arr);\\r\\n    const diffArr = arr.map((a) => (a - mu) ** 2);\\r\\n    return Math.sqrt(sum(diffArr) / (arr.length - 1));\\r\\n}\\r\\nfunction quantile(arr, q) {\\r\\n    const sorted = asc(arr);\\r\\n    const pos = (sorted.length - 1) * q;\\r\\n    const base = Math.floor(pos);\\r\\n    const rest = pos - base;\\r\\n    if (sorted[base + 1] != undefined) {\\r\\n        return sorted[base] + rest * (sorted[base + 1] - sorted[base]);\\r\\n    }\\r\\n    else if (sorted[base] != undefined) {\\r\\n        return sorted[base];\\r\\n    }\\r\\n    else {\\r\\n        return 0;\\r\\n    }\\r\\n}\\r\\nfunction markerPosition(x) {\\r\\n    let position = Math.log10(x) * 125 - 300;\\r\\n    if (position < 0) {\\r\\n        return 0;\\r\\n    }\\r\\n    else if (position > 100) {\\r\\n        return 100;\\r\\n    }\\r\\n    else {\\r\\n        return position;\\r\\n    }\\r\\n}\\r\\nfunction setMarkerPosition(median) {\\r\\n    let position = markerPosition(median);\\r\\n    marker.style.left = `${position}%`;\\r\\n}\\r\\nfunction build() {\\r\\n    let responseTimes = [];\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        responseTimes.push(data[i].response_time);\\r\\n    }\\r\\n    LQ = quantile(responseTimes, 0.25);\\r\\n    median = quantile(responseTimes, 0.5);\\r\\n    UQ = quantile(responseTimes, 0.75);\\r\\n    setMarkerPosition(median);\\r\\n}\\r\\nlet median;\\r\\nlet LQ;\\r\\nlet UQ;\\r\\nlet marker;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && build();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">\\r\\n    Response times <span class=\\\"milliseconds\\\">(ms)</span>\\r\\n  </div>\\r\\n  <div class=\\\"values\\\">\\r\\n    <div class=\\\"value lower-quartile\\\">{LQ}</div>\\r\\n    <div class=\\\"value median\\\">{median}</div>\\r\\n    <div class=\\\"value upper-quartile\\\">{UQ}</div>\\r\\n  </div>\\r\\n  <div class=\\\"labels\\\">\\r\\n    <div class=\\\"label\\\">25%</div>\\r\\n    <div class=\\\"label\\\">Median</div>\\r\\n    <div class=\\\"label\\\">75%</div>\\r\\n  </div>\\r\\n  <div class=\\\"bar\\\">\\r\\n    <div class=\\\"bar-green\\\" />\\r\\n    <div class=\\\"bar-yellow\\\" />\\r\\n    <div class=\\\"bar-red\\\" />\\r\\n    <div class=\\\"marker\\\" bind:this={marker} />\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .values {\\r\\n    display: flex;\\r\\n    color: var(--highlight);\\r\\n    font-size: 1.8em;\\r\\n    font-weight: 700;\\r\\n  }\\r\\n  .values,\\r\\n  .labels {\\r\\n    margin: 0 0.5rem;\\r\\n  }\\r\\n  .value {\\r\\n    flex: 1;\\r\\n    font-size: 1.1em;\\r\\n    padding: 20px 20px 4px;\\r\\n  }\\r\\n  .labels {\\r\\n    display: flex;\\r\\n    font-size: 0.8em;\\r\\n    color: var(--dim-text);\\r\\n  }\\r\\n  .label {\\r\\n    flex: 1;\\r\\n  }\\r\\n\\r\\n  .milliseconds {\\r\\n    color: var(--dim-text);\\r\\n    font-size: 0.8em;\\r\\n    margin-left: 4px;\\r\\n  }\\r\\n\\r\\n  .median {\\r\\n    font-size: 1em;\\r\\n  }\\r\\n  .upper-quartile,\\r\\n  .lower-quartile {\\r\\n    font-size: 1em;\\r\\n    padding-bottom: 0;\\r\\n  }\\r\\n\\r\\n  .bar {\\r\\n    padding: 20px 0 20px;\\r\\n    display: flex;\\r\\n    height: 30px;\\r\\n    width: 85%;\\r\\n    margin: auto;\\r\\n    align-items: center;\\r\\n    position: relative;\\r\\n  }\\r\\n  .bar-green {\\r\\n    background: var(--highlight);\\r\\n    width: 75%;\\r\\n    height: 10px;\\r\\n    border-radius: 2px 0 0 2px;\\r\\n  }\\r\\n  .bar-yellow {\\r\\n    width: 15%;\\r\\n    height: 10px;\\r\\n    background: rgb(235, 235, 129);\\r\\n  }\\r\\n  .bar-red {\\r\\n    width: 20%;\\r\\n    height: 10px;\\r\\n    border-radius: 0 2px 2px 0;\\r\\n    background: rgb(228, 97, 97);\\r\\n  }\\r\\n  .marker {\\r\\n    position: absolute;\\r\\n    height: 30px;\\r\\n    width: 5px;\\r\\n    background: white;\\r\\n    border-radius: 2px;\\r\\n    left: 0; /* Changed during runtime to reflect median */\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAwFE,OAAO,cAAC,CAAC,AACP,OAAO,CAAE,IAAI,CACb,KAAK,CAAE,IAAI,WAAW,CAAC,CACvB,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AACD,qBAAO,CACP,OAAO,cAAC,CAAC,AACP,MAAM,CAAE,CAAC,CAAC,MAAM,AAClB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,IAAI,CAAE,CAAC,CACP,SAAS,CAAE,KAAK,CAChB,OAAO,CAAE,IAAI,CAAC,IAAI,CAAC,GAAG,AACxB,CAAC,AACD,OAAO,cAAC,CAAC,AACP,OAAO,CAAE,IAAI,CACb,SAAS,CAAE,KAAK,CAChB,KAAK,CAAE,IAAI,UAAU,CAAC,AACxB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,IAAI,CAAE,CAAC,AACT,CAAC,AAED,aAAa,cAAC,CAAC,AACb,KAAK,CAAE,IAAI,UAAU,CAAC,CACtB,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AAED,OAAO,cAAC,CAAC,AACP,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,6BAAe,CACf,eAAe,cAAC,CAAC,AACf,SAAS,CAAE,GAAG,CACd,cAAc,CAAE,CAAC,AACnB,CAAC,AAED,IAAI,cAAC,CAAC,AACJ,OAAO,CAAE,IAAI,CAAC,CAAC,CAAC,IAAI,CACpB,OAAO,CAAE,IAAI,CACb,MAAM,CAAE,IAAI,CACZ,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,WAAW,CAAE,MAAM,CACnB,QAAQ,CAAE,QAAQ,AACpB,CAAC,AACD,UAAU,cAAC,CAAC,AACV,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,AAC5B,CAAC,AACD,WAAW,cAAC,CAAC,AACX,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,UAAU,CAAE,IAAI,GAAG,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAAC,AAChC,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,CAAC,CAAC,GAAG,CAAC,GAAG,CAAC,CAAC,CAC1B,UAAU,CAAE,IAAI,GAAG,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,AAC9B,CAAC,AACD,OAAO,cAAC,CAAC,AACP,QAAQ,CAAE,QAAQ,CAClB,MAAM,CAAE,IAAI,CACZ,KAAK,CAAE,GAAG,CACV,UAAU,CAAE,KAAK,CACjB,aAAa,CAAE,GAAG,CAClB,IAAI,CAAE,CAAC,AACT,CAAC\"}"
};

function markerPosition(x) {
	let position = Math.log10(x) * 125 - 300;

	if (position < 0) {
		return 0;
	} else if (position > 100) {
		return 100;
	} else {
		return position;
	}
}

const ResponseTimes = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	const asc = arr => arr.sort((a, b) => a - b);

	function quantile(arr, q) {
		const sorted = asc(arr);
		const pos = (sorted.length - 1) * q;
		const base = Math.floor(pos);
		const rest = pos - base;

		if (sorted[base + 1] != undefined) {
			return sorted[base] + rest * (sorted[base + 1] - sorted[base]);
		} else if (sorted[base] != undefined) {
			return sorted[base];
		} else {
			return 0;
		}
	}

	function setMarkerPosition(median) {
		let position = markerPosition(median);
		marker.style.left = `${position}%`;
	}

	function build() {
		let responseTimes = [];

		for (let i = 0; i < data.length; i++) {
			responseTimes.push(data[i].response_time);
		}

		LQ = quantile(responseTimes, 0.25);
		median = quantile(responseTimes, 0.5);
		UQ = quantile(responseTimes, 0.75);
		setMarkerPosition(median);
	}

	let median;
	let LQ;
	let UQ;
	let marker;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$g);
	data && mounted && build();

	return `<div class="${"card"}"><div class="${"card-title"}">Response times <span class="${"milliseconds svelte-kx5j01"}">(ms)</span></div>
  <div class="${"values svelte-kx5j01"}"><div class="${"value lower-quartile svelte-kx5j01"}">${escape(LQ)}</div>
    <div class="${"value median svelte-kx5j01"}">${escape(median)}</div>
    <div class="${"value upper-quartile svelte-kx5j01"}">${escape(UQ)}</div></div>
  <div class="${"labels svelte-kx5j01"}"><div class="${"label svelte-kx5j01"}">25%</div>
    <div class="${"label svelte-kx5j01"}">Median</div>
    <div class="${"label svelte-kx5j01"}">75%</div></div>
  <div class="${"bar svelte-kx5j01"}"><div class="${"bar-green svelte-kx5j01"}"></div>
    <div class="${"bar-yellow svelte-kx5j01"}"></div>
    <div class="${"bar-red svelte-kx5j01"}"></div>
    <div class="${"marker svelte-kx5j01"}"${add_attribute("this", marker, 0)}></div></div>
</div>`;
});

/* src\components\dashboard\Endpoints.svelte generated by Svelte v3.53.1 */

const css$f = {
	code: ".card.svelte-1l35be3{min-height:361px}.endpoints.svelte-1l35be3{margin:0.9em 20px 0.6em}.endpoint.svelte-1l35be3{border-radius:3px;margin:5px 0;color:var(--light-background);text-align:left;position:relative;font-size:0.85em}.endpoint-label.svelte-1l35be3{display:flex}.path.svelte-1l35be3,.count.svelte-1l35be3{padding:3px 15px}.count.svelte-1l35be3{margin-left:auto}.path.svelte-1l35be3{flex-grow:1;white-space:nowrap}.endpoint-container.svelte-1l35be3{display:flex}.external-label.svelte-1l35be3{padding:3px 15px;left:40px;top:0;margin:5px 0;color:#707070;display:none;font-size:0.85em}",
	map: "{\"version\":3,\"file\":\"Endpoints.svelte\",\"sources\":[\"Endpoints.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nlet methodMap = [\\\"GET\\\", \\\"POST\\\"];\\r\\nfunction endpointFreq() {\\r\\n    let freq = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let endpointID = data[i].path + data[i].status;\\r\\n        if (!(endpointID in freq)) {\\r\\n            freq[endpointID] = {\\r\\n                path: `${methodMap[data[i].method]}  ${data[i].path}`,\\r\\n                status: data[i].status,\\r\\n                count: 0,\\r\\n            };\\r\\n        }\\r\\n        freq[endpointID].count++;\\r\\n    }\\r\\n    return freq;\\r\\n}\\r\\nfunction build() {\\r\\n    let freq = endpointFreq();\\r\\n    let freqArr = [];\\r\\n    maxCount = 0;\\r\\n    for (let endpointID in freq) {\\r\\n        freqArr.push(freq[endpointID]);\\r\\n        if (freq[endpointID].count > maxCount) {\\r\\n            maxCount = freq[endpointID].count;\\r\\n        }\\r\\n    }\\r\\n    freqArr.sort((a, b) => {\\r\\n        return b.count - a.count;\\r\\n    });\\r\\n    endpoints = freqArr;\\r\\n    setTimeout(setEndpointLabels, 50);\\r\\n}\\r\\nfunction setEndpointLabelVisibility(idx) {\\r\\n    let endpoint = document.getElementById(`endpoint-label-${idx}`);\\r\\n    let endpointPath = document.getElementById(`endpoint-path-${idx}`);\\r\\n    let endpointCount = document.getElementById(`endpoint-count-${idx}`);\\r\\n    let externalLabel = document.getElementById(`external-label-${idx}`);\\r\\n    if (endpoint.clientWidth <\\r\\n        endpointPath.clientWidth + endpointCount.clientWidth) {\\r\\n        externalLabel.style.display = \\\"flex\\\";\\r\\n        endpointPath.style.display = \\\"none\\\";\\r\\n    }\\r\\n    if (endpoint.clientWidth < endpointCount.clientWidth) {\\r\\n        endpointCount.style.display = \\\"none\\\";\\r\\n    }\\r\\n}\\r\\nfunction setEndpointLabels() {\\r\\n    for (let i = 0; i < endpoints.length; i++) {\\r\\n        setEndpointLabelVisibility(i);\\r\\n    }\\r\\n}\\r\\nlet endpoints;\\r\\nlet maxCount;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && build();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">Endpoints</div>\\r\\n  {#if endpoints != undefined}\\r\\n    <div class=\\\"endpoints\\\">\\r\\n      {#each endpoints as endpoint, i}\\r\\n        <div class=\\\"endpoint-container\\\">\\r\\n          <div\\r\\n            class=\\\"endpoint\\\"\\r\\n            id=\\\"endpoint-{i}\\\"\\r\\n            title={endpoint.count}\\r\\n            style=\\\"width: {(endpoint.count / maxCount) *\\r\\n              100}%; background: {endpoint.status >= 200 &&\\r\\n            endpoint.status <= 299\\r\\n              ? 'var(--highlight)'\\r\\n              : '#e46161'}\\\"\\r\\n          >\\r\\n            <div class=\\\"endpoint-label\\\" id=\\\"endpoint-label-{i}\\\">\\r\\n              <div class=\\\"path\\\" id=\\\"endpoint-path-{i}\\\">\\r\\n                {endpoint.path}\\r\\n              </div>\\r\\n              <div class=\\\"count\\\" id=\\\"endpoint-count-{i}\\\">{endpoint.count}</div>\\r\\n            </div>\\r\\n          </div>\\r\\n          <div class=\\\"external-label\\\" id=\\\"external-label-{i}\\\">\\r\\n            <div class=\\\"external-label-path\\\">\\r\\n              {endpoint.path}\\r\\n            </div>\\r\\n          </div>\\r\\n        </div>\\r\\n      {/each}\\r\\n    </div>\\r\\n  {/if}\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    min-height: 361px;\\r\\n  }\\r\\n  .endpoints {\\r\\n    margin: 0.9em 20px 0.6em;\\r\\n  }\\r\\n  .endpoint {\\r\\n    border-radius: 3px;\\r\\n    margin: 5px 0;\\r\\n    color: var(--light-background);\\r\\n    text-align: left;\\r\\n    position: relative;\\r\\n    font-size: 0.85em;\\r\\n  }\\r\\n  .endpoint-label {\\r\\n    display: flex;\\r\\n  }\\r\\n  .path,\\r\\n  .count {\\r\\n    padding: 3px 15px;\\r\\n  }\\r\\n  .count {\\r\\n    margin-left: auto;\\r\\n  }\\r\\n  .path {\\r\\n    flex-grow: 1;\\r\\n    white-space: nowrap;\\r\\n  }\\r\\n  .endpoint-container {\\r\\n    display: flex;\\r\\n  }\\r\\n  .external-label {\\r\\n    padding: 3px 15px;\\r\\n    left: 40px;\\r\\n    top: 0;\\r\\n    margin: 5px 0;\\r\\n    color: #707070;\\r\\n    display: none;\\r\\n    font-size: 0.85em;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAiGE,KAAK,eAAC,CAAC,AACL,UAAU,CAAE,KAAK,AACnB,CAAC,AACD,UAAU,eAAC,CAAC,AACV,MAAM,CAAE,KAAK,CAAC,IAAI,CAAC,KAAK,AAC1B,CAAC,AACD,SAAS,eAAC,CAAC,AACT,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,GAAG,CAAC,CAAC,CACb,KAAK,CAAE,IAAI,kBAAkB,CAAC,CAC9B,UAAU,CAAE,IAAI,CAChB,QAAQ,CAAE,QAAQ,CAClB,SAAS,CAAE,MAAM,AACnB,CAAC,AACD,eAAe,eAAC,CAAC,AACf,OAAO,CAAE,IAAI,AACf,CAAC,AACD,oBAAK,CACL,MAAM,eAAC,CAAC,AACN,OAAO,CAAE,GAAG,CAAC,IAAI,AACnB,CAAC,AACD,MAAM,eAAC,CAAC,AACN,WAAW,CAAE,IAAI,AACnB,CAAC,AACD,KAAK,eAAC,CAAC,AACL,SAAS,CAAE,CAAC,CACZ,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,mBAAmB,eAAC,CAAC,AACnB,OAAO,CAAE,IAAI,AACf,CAAC,AACD,eAAe,eAAC,CAAC,AACf,OAAO,CAAE,GAAG,CAAC,IAAI,CACjB,IAAI,CAAE,IAAI,CACV,GAAG,CAAE,CAAC,CACN,MAAM,CAAE,GAAG,CAAC,CAAC,CACb,KAAK,CAAE,OAAO,CACd,OAAO,CAAE,IAAI,CACb,SAAS,CAAE,MAAM,AACnB,CAAC\"}"
};

function setEndpointLabelVisibility(idx) {
	let endpoint = document.getElementById(`endpoint-label-${idx}`);
	let endpointPath = document.getElementById(`endpoint-path-${idx}`);
	let endpointCount = document.getElementById(`endpoint-count-${idx}`);
	let externalLabel = document.getElementById(`external-label-${idx}`);

	if (endpoint.clientWidth < endpointPath.clientWidth + endpointCount.clientWidth) {
		externalLabel.style.display = "flex";
		endpointPath.style.display = "none";
	}

	if (endpoint.clientWidth < endpointCount.clientWidth) {
		endpointCount.style.display = "none";
	}
}

const Endpoints = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let methodMap = ["GET", "POST"];

	function endpointFreq() {
		let freq = {};

		for (let i = 0; i < data.length; i++) {
			let endpointID = data[i].path + data[i].status;

			if (!(endpointID in freq)) {
				freq[endpointID] = {
					path: `${methodMap[data[i].method]}  ${data[i].path}`,
					status: data[i].status,
					count: 0
				};
			}

			freq[endpointID].count++;
		}

		return freq;
	}

	function build() {
		let freq = endpointFreq();
		let freqArr = [];
		maxCount = 0;

		for (let endpointID in freq) {
			freqArr.push(freq[endpointID]);

			if (freq[endpointID].count > maxCount) {
				maxCount = freq[endpointID].count;
			}
		}

		freqArr.sort((a, b) => {
			return b.count - a.count;
		});

		endpoints = freqArr;
		setTimeout(setEndpointLabels, 50);
	}

	function setEndpointLabels() {
		for (let i = 0; i < endpoints.length; i++) {
			setEndpointLabelVisibility(i);
		}
	}

	let endpoints;
	let maxCount;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$f);
	data && mounted && build();

	return `<div class="${"card svelte-1l35be3"}"><div class="${"card-title"}">Endpoints</div>
  ${endpoints != undefined
	? `<div class="${"endpoints svelte-1l35be3"}">${each(endpoints, (endpoint, i) => {
			return `<div class="${"endpoint-container svelte-1l35be3"}"><div class="${"endpoint svelte-1l35be3"}" id="${"endpoint-" + escape(i, true)}"${add_attribute("title", endpoint.count, 0)} style="${"width: " + escape(endpoint.count / maxCount * 100, true) + "%; background: " + escape(
				endpoint.status >= 200 && endpoint.status <= 299
				? 'var(--highlight)'
				: '#e46161',
				true
			)}"><div class="${"endpoint-label svelte-1l35be3"}" id="${"endpoint-label-" + escape(i, true)}"><div class="${"path svelte-1l35be3"}" id="${"endpoint-path-" + escape(i, true)}">${escape(endpoint.path)}</div>
              <div class="${"count svelte-1l35be3"}" id="${"endpoint-count-" + escape(i, true)}">${escape(endpoint.count)}</div>
            </div></div>
          <div class="${"external-label svelte-1l35be3"}" id="${"external-label-" + escape(i, true)}"><div class="${"external-label-path"}">${escape(endpoint.path)}
            </div></div>
        </div>`;
		})}</div>`
	: ``}
</div>`;
});

/* src\components\dashboard\SuccessRate.svelte generated by Svelte v3.53.1 */

const css$e = {
	code: ".card.svelte-1vzzb7c{width:calc(200px - 1em);margin:0 0 2em 1em}.value.svelte-1vzzb7c{margin:20px 0;font-size:1.8em;font-weight:600;color:var(--yellow)}",
	map: "{\"version\":3,\"file\":\"SuccessRate.svelte\",\"sources\":[\"SuccessRate.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction build() {\\r\\n    let totalRequests = 0;\\r\\n    let successfulRequests = 0;\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        if (data[i].status >= 200 && data[i].status <= 299) {\\r\\n            successfulRequests++;\\r\\n        }\\r\\n        totalRequests++;\\r\\n    }\\r\\n    if (totalRequests > 0) {\\r\\n        successRate = (successfulRequests / totalRequests) * 100;\\r\\n    }\\r\\n    else {\\r\\n        successRate = 100;\\r\\n    }\\r\\n}\\r\\nlet mounted = false;\\r\\nlet successRate;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && build();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Last week\\\">\\r\\n  <div class=\\\"card-title\\\">Success rate</div>\\r\\n  {#if successRate != undefined}\\r\\n    <div\\r\\n      class=\\\"value\\\"\\r\\n      style=\\\"color: {successRate <= 75 ? 'var(--red)' : ''}{successRate > 75 &&\\r\\n      successRate < 90\\r\\n        ? 'var(--yellow)'\\r\\n        : ''}{successRate >= 90 ? 'var(--highlight)' : ''}\\\"\\r\\n    >\\r\\n      {successRate.toFixed(1)}%\\r\\n    </div>\\r\\n  {/if}\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    width: calc(200px - 1em);\\r\\n    margin: 0 0 2em 1em;\\r\\n  }\\r\\n\\r\\n  .value {\\r\\n    margin: 20px 0;\\r\\n    font-size: 1.8em;\\r\\n    font-weight: 600;\\r\\n    color: var(--yellow);\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AA0CE,KAAK,eAAC,CAAC,AACL,KAAK,CAAE,KAAK,KAAK,CAAC,CAAC,CAAC,GAAG,CAAC,CACxB,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CAAC,GAAG,AACrB,CAAC,AAED,MAAM,eAAC,CAAC,AACN,MAAM,CAAE,IAAI,CAAC,CAAC,CACd,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,CAChB,KAAK,CAAE,IAAI,QAAQ,CAAC,AACtB,CAAC\"}"
};

const SuccessRate = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function build() {
		let totalRequests = 0;
		let successfulRequests = 0;

		for (let i = 0; i < data.length; i++) {
			if (data[i].status >= 200 && data[i].status <= 299) {
				successfulRequests++;
			}

			totalRequests++;
		}

		if (totalRequests > 0) {
			successRate = successfulRequests / totalRequests * 100;
		} else {
			successRate = 100;
		}
	}

	let mounted = false;
	let successRate;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$e);
	data && mounted && build();

	return `<div class="${"card svelte-1vzzb7c"}" title="${"Last week"}"><div class="${"card-title"}">Success rate</div>
  ${successRate != undefined
	? `<div class="${"value svelte-1vzzb7c"}" style="${"color: " + escape(successRate <= 75 ? 'var(--red)' : '', true) + escape(
			successRate > 75 && successRate < 90
			? 'var(--yellow)'
			: '',
			true
		) + escape(successRate >= 90 ? 'var(--highlight)' : '', true)}">${escape(successRate.toFixed(1))}%
    </div>`
	: ``}
</div>`;
});

/* src\components\dashboard\activity\ActivityRequests.svelte generated by Svelte v3.53.1 */

const ActivityRequests = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function defaultLayout() {
		let periodAgo = new Date();
		let days = periodToDays$1(period);

		if (days != null) {
			periodAgo.setDate(periodAgo.getDate() - days);
		} else {
			periodAgo = null;
		}

		let tomorrow = new Date();
		tomorrow.setDate(tomorrow.getDate());

		return {
			title: false,
			autosize: true,
			margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
			hovermode: "closest",
			plot_bgcolor: "transparent",
			paper_bgcolor: "transparent",
			height: 169,
			yaxis: {
				title: { text: "Requests" },
				gridcolor: "gray",
				showgrid: false
			},
			xaxis: {
				title: { text: "Date" },
				fixedrange: true,
				range: [periodAgo, tomorrow],
				visible: false
			},
			dragmode: false
		};
	}

	function bars() {
		let requestFreq = {};
		let days = periodToDays$1(period);

		if (days) {
			for (let i = 0; i < days; i++) {
				let date = new Date();
				date.setHours(0, 0, 0, 0);
				date.setDate(date.getDate() - i);

				// @ts-ignore
				requestFreq[date] = 0;
			}
		}

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);
			date.setHours(0, 0, 0, 0);

			// @ts-ignore
			if (!(date in requestFreq)) {
				// @ts-ignore
				requestFreq[date] = 0;
			}

			// @ts-ignore
			requestFreq[date]++;
		}

		let requestFreqArr = [];

		for (let date in requestFreq) {
			requestFreqArr.push([new Date(date), requestFreq[date]]);
		}

		requestFreqArr.sort((a, b) => {
			return a[0] - b[0];
		});

		let dates = [];
		let requests = [];

		for (let i = 0; i < requestFreqArr.length; i++) {
			dates.push(requestFreqArr[i][0]);
			requests.push(requestFreqArr[i][1]);
		}

		return [
			{
				x: dates,
				y: requests,
				type: "bar",
				marker: { color: "#3fcf8e" },
				hovertemplate: `<b>%{y} requests</b><br>%{x|%d %b %Y}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	function buildPlotData() {
		return {
			data: bars(),
			layout: defaultLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	let plotDiv;

	function genPlot() {
		let plotData = buildPlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data, period } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.period === void 0 && $$bindings.period && period !== void 0) $$bindings.period(period);
	data && mounted && genPlot();
	return `<div id="${"plotly"}"><div id="${"plotDiv"}"${add_attribute("this", plotDiv, 0)}></div></div>`;
});

/* src\components\dashboard\activity\ActivityResponseTime.svelte generated by Svelte v3.53.1 */

const ActivityResponseTime = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function defaultLayout() {
		let periodAgo = new Date();
		let days = periodToDays$1(period);

		if (days != null) {
			periodAgo.setDate(periodAgo.getDate() - days);
		} else {
			periodAgo = null;
		}

		let tomorrow = new Date();
		tomorrow.setDate(tomorrow.getDate());

		return {
			title: false,
			autosize: true,
			margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
			hovermode: "closest",
			plot_bgcolor: "transparent",
			paper_bgcolor: "transparent",
			height: 170,
			yaxis: {
				title: { text: "Response time (ms)" },
				gridcolor: "gray",
				showgrid: false,
				fixedrange: true
			},
			xaxis: {
				title: { text: "Date" },
				showgrid: false,
				fixedrange: true,
				range: [periodAgo, tomorrow],
				visible: false
			},
			dragmode: false
		};
	}

	function bars() {
		let responseTimes = {};
		let days = periodToDays$1(period);

		if (days) {
			for (let i = 0; i < days; i++) {
				let date = new Date();
				date.setHours(0, 0, 0, 0);
				date.setDate(date.getDate() - i);

				// @ts-ignore
				responseTimes[date] = { total: 0, count: 0 };
			}
		}

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);
			date.setHours(0, 0, 0, 0);

			// @ts-ignore
			if (!(date in responseTimes)) {
				// @ts-ignore
				responseTimes[date] = { total: 0, count: 0 };
			}

			// @ts-ignore
			responseTimes[date].total += data[i].response_time;

			// @ts-ignore
			responseTimes[date].count++;
		}

		let requestFreqArr = [];

		for (let date in responseTimes) {
			let point = [new Date(date), 0];

			if (responseTimes[date].count > 0) {
				point[1] = responseTimes[date].total / responseTimes[date].count;
			}

			requestFreqArr.push(point);
		}

		requestFreqArr.sort((a, b) => {
			return a[0] - b[0];
		});

		let dates = [];
		let requests = [];

		for (let i = 0; i < requestFreqArr.length; i++) {
			dates.push(requestFreqArr[i][0]);
			requests.push(requestFreqArr[i][1]);
		}

		return [
			{
				x: dates,
				y: requests,
				type: "lines",
				marker: { color: "#707070" },
				// fill: "tonexty",
				hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y}</b><extra></extra>`,
				showlegend: false
			},
			{
				x: dates,
				y: Array(requests.length).fill(Math.max(Math.min(...requests) - 5, 0)),
				type: "lines",
				marker: { color: "transparent" },
				fill: "tonexty",
				fillcolor: "#4A4A4A",
				hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	function buildPlotData() {
		return {
			data: bars(),
			layout: defaultLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot() {
		let plotData = buildPlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let plotDiv;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data, period } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.period === void 0 && $$bindings.period && period !== void 0) $$bindings.period(period);
	data && mounted && genPlot();
	return `<div id="${"plotly"}"><div id="${"plotDiv"}"${add_attribute("this", plotDiv, 0)}></div></div>`;
});

/* src\components\dashboard\activity\ActivitySuccessRate.svelte generated by Svelte v3.53.1 */

const css$d = {
	code: ".errors.svelte-13xx9v{display:flex;margin-top:8px;margin:0 10px 0 40px}.error.svelte-13xx9v{background:var(--highlight);flex:1;height:40px;margin:0 0.1%;border-radius:1px}.success-rate-container.svelte-13xx9v{text-align:left;font-size:0.9em;color:#707070}.success-rate-title.svelte-13xx9v{margin:0 0 4px 43px}.success-rate-container.svelte-13xx9v{margin:1.5em 2.5em 2em}.level-0.svelte-13xx9v{background:rgb(40, 40, 40)}.level-1.svelte-13xx9v{background:#e46161}.level-2.svelte-13xx9v{background:#f18359}.level-3.svelte-13xx9v{background:#f5a65a}.level-4.svelte-13xx9v{background:#f3c966}.level-5.svelte-13xx9v{background:#ebeb81}.level-6.svelte-13xx9v{background:#c7e57d}.level-7.svelte-13xx9v{background:#a1df7e}.level-8.svelte-13xx9v{background:#77d884}.level-9.svelte-13xx9v{background:#3fcf8e}",
	map: "{\"version\":3,\"file\":\"ActivitySuccessRate.svelte\",\"sources\":[\"ActivitySuccessRate.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nimport periodToDays from \\\"../../../lib/period\\\";\\r\\nfunction daysAgo(date) {\\r\\n    let now = new Date();\\r\\n    return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));\\r\\n}\\r\\nfunction setSuccessRate() {\\r\\n    let success = {};\\r\\n    let minDate = Number.POSITIVE_INFINITY;\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        date.setHours(0, 0, 0, 0);\\r\\n        // @ts-ignore\\r\\n        if (!(date in success)) {\\r\\n            // @ts-ignore\\r\\n            success[date] = { total: 0, successful: 0 };\\r\\n        }\\r\\n        if (data[i].status >= 200 && data[i].status <= 299) {\\r\\n            // @ts-ignore\\r\\n            success[date].successful++;\\r\\n        }\\r\\n        // @ts-ignore\\r\\n        success[date].total++;\\r\\n        if (date < minDate) {\\r\\n            minDate = date;\\r\\n        }\\r\\n    }\\r\\n    let days = periodToDays(period);\\r\\n    if (days == null) {\\r\\n        days = daysAgo(minDate);\\r\\n    }\\r\\n    let successArr = new Array(days).fill(-0.1); // -0.1 -> 0\\r\\n    for (let date in success) {\\r\\n        let idx = daysAgo(new Date(date));\\r\\n        successArr[successArr.length - 1 - idx] =\\r\\n            success[date].successful / success[date].total;\\r\\n    }\\r\\n    successRate = successArr;\\r\\n}\\r\\nfunction build() {\\r\\n    setSuccessRate();\\r\\n}\\r\\nlet successRate;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: successRate;\\r\\n$: data && mounted && build();\\r\\nexport let data, period;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"success-rate-container\\\">\\r\\n  {#if successRate != undefined}\\r\\n    <div class=\\\"success-rate-title\\\">Success rate</div>\\r\\n    <div class=\\\"errors\\\">\\r\\n      {#each successRate as value}\\r\\n        <div\\r\\n          class=\\\"error level-{Math.floor(value * 10) + 1}\\\"\\r\\n          title={value >= 0 ? (value * 100).toFixed(1) + \\\"%\\\" : \\\"No requests\\\"}\\r\\n        />\\r\\n      {/each}\\r\\n    </div>\\r\\n  {/if}\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .errors {\\r\\n    display: flex;\\r\\n    margin-top: 8px;\\r\\n    margin: 0 10px 0 40px;\\r\\n  }\\r\\n  .error {\\r\\n    background: var(--highlight);\\r\\n    flex: 1;\\r\\n    height: 40px;\\r\\n    margin: 0 0.1%;\\r\\n    border-radius: 1px;\\r\\n  }\\r\\n  .success-rate-container {\\r\\n    text-align: left;\\r\\n    font-size: 0.9em;\\r\\n    color: #707070;\\r\\n  }\\r\\n  .success-rate-title {\\r\\n    margin: 0 0 4px 43px;\\r\\n  }\\r\\n  .success-rate-container {\\r\\n    margin: 1.5em 2.5em 2em;\\r\\n  }\\r\\n  .level-0 {\\r\\n    background: rgb(40, 40, 40);\\r\\n  }\\r\\n  .level-1 {\\r\\n    background: #e46161;\\r\\n  }\\r\\n  .level-2 {\\r\\n    background: #f18359;\\r\\n  }\\r\\n  .level-3 {\\r\\n    background: #f5a65a;\\r\\n  }\\r\\n  .level-4 {\\r\\n    background: #f3c966;\\r\\n  }\\r\\n  .level-5 {\\r\\n    background: #ebeb81;\\r\\n  }\\r\\n  .level-6 {\\r\\n    background: #c7e57d;\\r\\n  }\\r\\n  .level-7 {\\r\\n    background: #a1df7e;\\r\\n  }\\r\\n  .level-8 {\\r\\n    background: #77d884;\\r\\n  }\\r\\n  .level-9 {\\r\\n    background: #3fcf8e;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAmEE,OAAO,cAAC,CAAC,AACP,OAAO,CAAE,IAAI,CACb,UAAU,CAAE,GAAG,CACf,MAAM,CAAE,CAAC,CAAC,IAAI,CAAC,CAAC,CAAC,IAAI,AACvB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,IAAI,CAAE,CAAC,CACP,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,CAAC,CAAC,IAAI,CACd,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,UAAU,CAAE,IAAI,CAChB,SAAS,CAAE,KAAK,CAChB,KAAK,CAAE,OAAO,AAChB,CAAC,AACD,mBAAmB,cAAC,CAAC,AACnB,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CAAC,IAAI,AACtB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,MAAM,CAAE,KAAK,CAAC,KAAK,CAAC,GAAG,AACzB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,IAAI,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,AAC7B,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,AACrB,CAAC\"}"
};

function daysAgo(date) {
	let now = new Date();
	return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));
}

const ActivitySuccessRate = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function setSuccessRate() {
		let success = {};
		let minDate = Number.POSITIVE_INFINITY;

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);
			date.setHours(0, 0, 0, 0);

			// @ts-ignore
			if (!(date in success)) {
				// @ts-ignore
				success[date] = { total: 0, successful: 0 };
			}

			if (data[i].status >= 200 && data[i].status <= 299) {
				// @ts-ignore
				success[date].successful++;
			}

			// @ts-ignore
			success[date].total++;

			if (date < minDate) {
				minDate = date;
			}
		}

		let days = periodToDays$1(period);

		if (days == null) {
			days = daysAgo(minDate);
		}

		let successArr = new Array(days).fill(-0.1); // -0.1 -> 0

		for (let date in success) {
			let idx = daysAgo(new Date(date));
			successArr[successArr.length - 1 - idx] = success[date].successful / success[date].total;
		}

		successRate = successArr;
	}

	function build() {
		setSuccessRate();
	}

	let successRate;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data, period } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.period === void 0 && $$bindings.period && period !== void 0) $$bindings.period(period);
	$$result.css.add(css$d);

	data && mounted && build();

	return `<div class="${"success-rate-container svelte-13xx9v"}">${successRate != undefined
	? `<div class="${"success-rate-title svelte-13xx9v"}">Success rate</div>
    <div class="${"errors svelte-13xx9v"}">${each(successRate, value => {
			return `<div class="${"error level-" + escape(Math.floor(value * 10) + 1, true) + " svelte-13xx9v"}"${add_attribute(
				"title",
				value >= 0
				? (value * 100).toFixed(1) + "%"
				: "No requests",
				0
			)}></div>`;
		})}</div>`
	: ``}
</div>`;
});

/* src\components\dashboard\activity\Activity.svelte generated by Svelte v3.53.1 */

const css$c = {
	code: ".card.svelte-1snjwfx{margin:0;width:100%}",
	map: "{\"version\":3,\"file\":\"Activity.svelte\",\"sources\":[\"Activity.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import ActivityRequests from \\\"./ActivityRequests.svelte\\\";\\r\\nimport ActivityResponseTime from \\\"./ActivityResponseTime.svelte\\\";\\r\\nimport ActivitySuccessRate from \\\"./ActivitySuccessRate.svelte\\\";\\r\\nexport let data, period;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">Activity</div>\\r\\n  <ActivityRequests {data} {period} />\\r\\n  <ActivityResponseTime {data} {period} />\\r\\n  <ActivitySuccessRate {data} {period} />\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    margin: 0;\\r\\n    width: 100%;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAcE,KAAK,eAAC,CAAC,AACL,MAAM,CAAE,CAAC,CACT,KAAK,CAAE,IAAI,AACb,CAAC\"}"
};

const Activity = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let { data, period } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.period === void 0 && $$bindings.period && period !== void 0) $$bindings.period(period);
	$$result.css.add(css$c);

	return `<div class="${"card svelte-1snjwfx"}"><div class="${"card-title"}">Activity</div>
  ${validate_component(ActivityRequests, "ActivityRequests").$$render($$result, { data, period }, {}, {})}
  ${validate_component(ActivityResponseTime, "ActivityResponseTime").$$render($$result, { data, period }, {}, {})}
  ${validate_component(ActivitySuccessRate, "ActivitySuccessRate").$$render($$result, { data, period }, {}, {})}
</div>`;
});

/* src\components\dashboard\Version.svelte generated by Svelte v3.53.1 */

const css$b = {
	code: ".card.svelte-jecwjn{margin:2em 0 2em 2em;padding-bottom:1em;flex:1}#plotDiv.svelte-jecwjn{margin-right:20px}",
	map: "{\"version\":3,\"file\":\"Version.svelte\",\"sources\":[\"Version.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction setVersions() {\\r\\n    let v = new Set();\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let match = data[i].path.match(/[^a-z0-9](v\\\\d)[^a-z0-9]/i);\\r\\n        if (match) {\\r\\n            v.add(match[1]);\\r\\n        }\\r\\n    }\\r\\n    versions = v;\\r\\n    if (versions.size > 1) {\\r\\n        setTimeout(genPlot, 1000);\\r\\n    }\\r\\n}\\r\\nfunction versionPlotLayout() {\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 180,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Requests\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nlet colors = [\\r\\n    \\\"#3FCF8E\\\",\\r\\n    \\\"#5784BA\\\",\\r\\n    \\\"#EBEB81\\\",\\r\\n    \\\"#218B82\\\",\\r\\n    \\\"#FFD6A5\\\",\\r\\n    \\\"#F9968B\\\",\\r\\n    \\\"#B1A2CA\\\",\\r\\n    \\\"#E46161\\\", // Red\\r\\n];\\r\\nfunction pieChart() {\\r\\n    let versionCount = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let match = data[i].path.match(/[^a-z0-9](v\\\\d)[^a-z0-9]/i);\\r\\n        if (match) {\\r\\n            let version = match[0];\\r\\n            if (!(version in versionCount)) {\\r\\n                versionCount[version] = 0;\\r\\n            }\\r\\n            versionCount[version]++;\\r\\n        }\\r\\n    }\\r\\n    let versions = [];\\r\\n    let count = [];\\r\\n    for (let version in versionCount) {\\r\\n        versions.push(version);\\r\\n        count.push(versionCount[version]);\\r\\n    }\\r\\n    return [\\r\\n        {\\r\\n            values: count,\\r\\n            labels: versions,\\r\\n            type: \\\"pie\\\",\\r\\n            marker: {\\r\\n                colors: colors,\\r\\n            },\\r\\n        },\\r\\n    ];\\r\\n}\\r\\nfunction versionPlotData() {\\r\\n    return {\\r\\n        data: pieChart(),\\r\\n        layout: versionPlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genPlot() {\\r\\n    let plotData = versionPlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\nlet versions;\\r\\nlet plotDiv;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && setVersions();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n{#if versions != undefined && versions.size > 1}\\r\\n  <div class=\\\"card\\\">\\r\\n    <div class=\\\"card-title\\\">Version</div>\\r\\n    <div id=\\\"plotly\\\">\\r\\n      <div id=\\\"plotDiv\\\" bind:this={plotDiv}>\\r\\n        <!-- Plotly chart will be drawn inside this DIV -->\\r\\n      </div>\\r\\n    </div>\\r\\n  </div>\\r\\n{/if}\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    margin: 2em 0 2em 2em;\\r\\n    padding-bottom: 1em;\\r\\n    flex: 1;\\r\\n  }\\r\\n  #plotDiv {\\r\\n    margin-right: 20px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAgHE,KAAK,cAAC,CAAC,AACL,MAAM,CAAE,GAAG,CAAC,CAAC,CAAC,GAAG,CAAC,GAAG,CACrB,cAAc,CAAE,GAAG,CACnB,IAAI,CAAE,CAAC,AACT,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,YAAY,CAAE,IAAI,AACpB,CAAC\"}"
};

function versionPlotLayout() {
	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
		hovermode: "closest",
		plot_bgcolor: "transparent",
		paper_bgcolor: "transparent",
		height: 180,
		yaxis: {
			title: { text: "Requests" },
			gridcolor: "gray",
			showgrid: false,
			fixedrange: true
		},
		xaxis: { visible: false },
		dragmode: false
	};
}

const Version = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function setVersions() {
		let v = new Set();

		for (let i = 0; i < data.length; i++) {
			let match = data[i].path.match(/[^a-z0-9](v\d)[^a-z0-9]/i);

			if (match) {
				v.add(match[1]);
			}
		}

		versions = v;

		if (versions.size > 1) {
			setTimeout(genPlot, 1000);
		}
	}

	let colors = [
		"#3FCF8E",
		"#5784BA",
		"#EBEB81",
		"#218B82",
		"#FFD6A5",
		"#F9968B",
		"#B1A2CA",
		"#E46161"
	]; // Red

	function pieChart() {
		let versionCount = {};

		for (let i = 0; i < data.length; i++) {
			let match = data[i].path.match(/[^a-z0-9](v\d)[^a-z0-9]/i);

			if (match) {
				let version = match[0];

				if (!(version in versionCount)) {
					versionCount[version] = 0;
				}

				versionCount[version]++;
			}
		}

		let versions = [];
		let count = [];

		for (let version in versionCount) {
			versions.push(version);
			count.push(versionCount[version]);
		}

		return [
			{
				values: count,
				labels: versions,
				type: "pie",
				marker: { colors }
			}
		];
	}

	function versionPlotData() {
		return {
			data: pieChart(),
			layout: versionPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot() {
		let plotData = versionPlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let versions;
	let plotDiv;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$b);
	data && mounted && setVersions();

	return `${versions != undefined && versions.size > 1
	? `<div class="${"card svelte-jecwjn"}"><div class="${"card-title"}">Version</div>
    <div id="${"plotly"}"><div id="${"plotDiv"}" class="${"svelte-jecwjn"}"${add_attribute("this", plotDiv, 0)}></div></div></div>`
	: ``}`;
});

/* src\components\dashboard\UsageTime.svelte generated by Svelte v3.53.1 */

const css$a = {
	code: ".card.svelte-ecbw89{width:100%;margin:0}",
	map: "{\"version\":3,\"file\":\"UsageTime.svelte\",\"sources\":[\"UsageTime.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction defaultLayout() {\\r\\n    return {\\r\\n        font: { size: 12 },\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 500,\\r\\n        margin: { r: 35, l: 70, t: 20, b: 50, pad: 0 },\\r\\n        polar: {\\r\\n            bargap: 0,\\r\\n            bgcolor: \\\"transparent\\\",\\r\\n            angularaxis: { direction: \\\"clockwise\\\", showgrid: false },\\r\\n            radialaxis: { gridcolor: \\\"#303030\\\" },\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction bars() {\\r\\n    let responseTimes = Array(24).fill(0);\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        let time = date.getHours();\\r\\n        // @ts-ignore\\r\\n        responseTimes[time]++;\\r\\n    }\\r\\n    let requestFreqArr = [];\\r\\n    for (let i = 0; i < 24; i++) {\\r\\n        let point = [i, responseTimes[i]];\\r\\n        requestFreqArr.push(point);\\r\\n    }\\r\\n    requestFreqArr.sort((a, b) => {\\r\\n        return a[0] - b[0];\\r\\n    });\\r\\n    let dates = [];\\r\\n    let requests = [];\\r\\n    for (let i = 0; i < requestFreqArr.length; i++) {\\r\\n        dates.push(requestFreqArr[i][0].toString() + \\\":00\\\");\\r\\n        requests.push(requestFreqArr[i][1]);\\r\\n    }\\r\\n    return [\\r\\n        {\\r\\n            r: requests,\\r\\n            theta: dates,\\r\\n            marker: { color: \\\"#3fcf8e\\\" },\\r\\n            type: \\\"barpolar\\\",\\r\\n            hovertemplate: `<b>%{r}</b> requests at <b>%{theta}</b><extra></extra>`,\\r\\n        },\\r\\n    ];\\r\\n}\\r\\nfunction buildPlotData() {\\r\\n    return {\\r\\n        data: bars(),\\r\\n        layout: defaultLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genPlot() {\\r\\n    let plotData = buildPlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\nlet plotDiv;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && genPlot();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">Usage time</div>\\r\\n  <div id=\\\"plotly\\\">\\r\\n    <div id=\\\"plotDiv\\\" bind:this={plotDiv}>\\r\\n      <!-- Plotly chart will be drawn inside this DIV -->\\r\\n    </div>\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    width: 100%;\\r\\n    margin: 0;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAkFE,KAAK,cAAC,CAAC,AACL,KAAK,CAAE,IAAI,CACX,MAAM,CAAE,CAAC,AACX,CAAC\"}"
};

function defaultLayout$1() {
	return {
		font: { size: 12 },
		paper_bgcolor: "transparent",
		height: 500,
		margin: { r: 35, l: 70, t: 20, b: 50, pad: 0 },
		polar: {
			bargap: 0,
			bgcolor: "transparent",
			angularaxis: { direction: "clockwise", showgrid: false },
			radialaxis: { gridcolor: "#303030" }
		}
	};
}

const UsageTime = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function bars() {
		let responseTimes = Array(24).fill(0);

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);
			let time = date.getHours();

			// @ts-ignore
			responseTimes[time]++;
		}

		let requestFreqArr = [];

		for (let i = 0; i < 24; i++) {
			let point = [i, responseTimes[i]];
			requestFreqArr.push(point);
		}

		requestFreqArr.sort((a, b) => {
			return a[0] - b[0];
		});

		let dates = [];
		let requests = [];

		for (let i = 0; i < requestFreqArr.length; i++) {
			dates.push(requestFreqArr[i][0].toString() + ":00");
			requests.push(requestFreqArr[i][1]);
		}

		return [
			{
				r: requests,
				theta: dates,
				marker: { color: "#3fcf8e" },
				type: "barpolar",
				hovertemplate: `<b>%{r}</b> requests at <b>%{theta}</b><extra></extra>`
			}
		];
	}

	function buildPlotData() {
		return {
			data: bars(),
			layout: defaultLayout$1(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot() {
		let plotData = buildPlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let plotDiv;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$a);
	data && mounted && genPlot();

	return `<div class="${"card svelte-ecbw89"}"><div class="${"card-title"}">Usage time</div>
  <div id="${"plotly"}"><div id="${"plotDiv"}"${add_attribute("this", plotDiv, 0)}></div></div>
</div>`;
});

/* src\components\dashboard\Growth.svelte generated by Svelte v3.53.1 */

const css$9 = {
	code: ".card.svelte-1v5417n{flex:1.2;margin:2em 1em 2em 0}.values.svelte-1v5417n{margin:30px 30px;display:flex}.tile-value.svelte-1v5417n{font-size:1.4em;margin-bottom:5px}.tile-label.svelte-1v5417n{font-size:0.8em}.tile.svelte-1v5417n{background:#282828;flex:1;padding:30px 10px;border-radius:6px;margin:10px}@media screen and (max-width: 1580px){.card.svelte-1v5417n{width:100%}}@media screen and (max-width: 1580px){.card.svelte-1v5417n{margin:2em 0 2em;width:100%}}",
	map: "{\"version\":3,\"file\":\"Growth.svelte\",\"sources\":[\"Growth.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction periodData(data) {\\r\\n    let period = { requests: 0, success: 0, responseTime: 0 };\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        // @ts-ignore\\r\\n        period.requests++;\\r\\n        if (data[i].status >= 200 && data[i].status <= 299) {\\r\\n            period.success++;\\r\\n        }\\r\\n        period.responseTime += data[i].response_time;\\r\\n    }\\r\\n    return period;\\r\\n}\\r\\nfunction buildWeek() {\\r\\n    let thisPeriod = periodData(data);\\r\\n    let lastPeriod = periodData(prevData);\\r\\n    let requestsChange = ((thisPeriod.requests + 1) / (lastPeriod.requests + 1)) * 100 - 100;\\r\\n    let successChange = (((thisPeriod.success + 1) / (thisPeriod.requests + 1) + 1) /\\r\\n        ((lastPeriod.success + 1) / (lastPeriod.requests + 1) + 1)) *\\r\\n        100 -\\r\\n        100;\\r\\n    let responseTimeChange = (((thisPeriod.responseTime + 1) / (thisPeriod.requests + 1) + 1) /\\r\\n        ((lastPeriod.responseTime + 1) / (lastPeriod.requests + 1) + 1)) *\\r\\n        100 -\\r\\n        100;\\r\\n    change = {\\r\\n        requests: requestsChange,\\r\\n        success: successChange,\\r\\n        responseTime: responseTimeChange,\\r\\n    };\\r\\n}\\r\\nlet change;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && buildWeek();\\r\\nexport let data, prevData;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">Growth</div>\\r\\n  <div class=\\\"values\\\">\\r\\n    {#if change != undefined}\\r\\n      <div class=\\\"tile\\\">\\r\\n        <div class=\\\"tile-value\\\">\\r\\n          <span\\r\\n            style=\\\"color: {change.requests > 0\\r\\n              ? 'var(--highlight)'\\r\\n              : 'var(--red)'}\\\"\\r\\n            >{change.requests > 0 ? \\\"+\\\" : \\\"\\\"}{change.requests.toFixed(1)}%</span\\r\\n          >\\r\\n        </div>\\r\\n        <div class=\\\"tile-label\\\">Requests</div>\\r\\n      </div>\\r\\n      <!-- {#if requestsChange != undefined}\\r\\n        <div class=\\\"tile\\\">\\r\\n          <div class=\\\"tile-value\\\">\\r\\n            <span\\r\\n              style=\\\"color: {requestsChange > 0\\r\\n                ? 'var(--highlight)'\\r\\n                : 'var(--red)'}\\\">+{requestsChange}%</span\\r\\n            >\\r\\n          </div>\\r\\n          <div class=\\\"tile-label\\\">Users</div>\\r\\n        </div>\\r\\n    {/if} -->\\r\\n      <div class=\\\"tile\\\">\\r\\n        <div class=\\\"tile-value\\\">\\r\\n          <span\\r\\n            style=\\\"color: {change.success > 0\\r\\n              ? 'var(--highlight)'\\r\\n              : 'var(--red)'}\\\"\\r\\n            >{change.success > 0 ? \\\"+\\\" : \\\"\\\"}{change.success.toFixed(1)}%</span\\r\\n          >\\r\\n        </div>\\r\\n        <div class=\\\"tile-label\\\">Success rate</div>\\r\\n      </div>\\r\\n      <div class=\\\"tile\\\">\\r\\n        <div class=\\\"tile-value\\\">\\r\\n          <span\\r\\n            style=\\\"color: {change.responseTime < 0\\r\\n              ? 'var(--highlight)'\\r\\n              : 'var(--red)'}\\\"\\r\\n            >{change.responseTime > 0 ? \\\"+\\\" : \\\"\\\"}{change.responseTime.toFixed(\\r\\n              1\\r\\n            )}%</span\\r\\n          >\\r\\n        </div>\\r\\n        <div class=\\\"tile-label\\\">Response time</div>\\r\\n      </div>\\r\\n    {/if}\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    /* width: 100%; */\\r\\n    flex: 1.2;\\r\\n    margin: 2em 1em 2em 0;\\r\\n  }\\r\\n  .values {\\r\\n    margin: 30px 30px;\\r\\n    display: flex;\\r\\n  }\\r\\n  .tile-value {\\r\\n    font-size: 1.4em;\\r\\n    margin-bottom: 5px;\\r\\n  }\\r\\n  .tile-label {\\r\\n    font-size: 0.8em;\\r\\n  }\\r\\n  .tile {\\r\\n    background: #282828;\\r\\n    /* border: 1px solid var(--highlight); */\\r\\n    flex: 1;\\r\\n    padding: 30px 10px;\\r\\n    border-radius: 6px;\\r\\n    margin: 10px;\\r\\n  }\\r\\n  @media screen and (max-width: 1580px) {\\r\\n    .card {\\r\\n      width: 100%;\\r\\n    }\\r\\n  }\\r\\n  @media screen and (max-width: 1580px) {\\r\\n    .card {\\r\\n      margin: 2em 0 2em;\\r\\n      width: 100%;\\r\\n    }\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAgGE,KAAK,eAAC,CAAC,AAEL,IAAI,CAAE,GAAG,CACT,MAAM,CAAE,GAAG,CAAC,GAAG,CAAC,GAAG,CAAC,CAAC,AACvB,CAAC,AACD,OAAO,eAAC,CAAC,AACP,MAAM,CAAE,IAAI,CAAC,IAAI,CACjB,OAAO,CAAE,IAAI,AACf,CAAC,AACD,WAAW,eAAC,CAAC,AACX,SAAS,CAAE,KAAK,CAChB,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,WAAW,eAAC,CAAC,AACX,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,KAAK,eAAC,CAAC,AACL,UAAU,CAAE,OAAO,CAEnB,IAAI,CAAE,CAAC,CACP,OAAO,CAAE,IAAI,CAAC,IAAI,CAClB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,AACd,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,KAAK,eAAC,CAAC,AACL,KAAK,CAAE,IAAI,AACb,CAAC,AACH,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,KAAK,eAAC,CAAC,AACL,MAAM,CAAE,GAAG,CAAC,CAAC,CAAC,GAAG,CACjB,KAAK,CAAE,IAAI,AACb,CAAC,AACH,CAAC\"}"
};

function periodData(data) {
	let period = { requests: 0, success: 0, responseTime: 0 };

	for (let i = 0; i < data.length; i++) {
		// @ts-ignore
		period.requests++;

		if (data[i].status >= 200 && data[i].status <= 299) {
			period.success++;
		}

		period.responseTime += data[i].response_time;
	}

	return period;
}

const Growth = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function buildWeek() {
		let thisPeriod = periodData(data);
		let lastPeriod = periodData(prevData);
		let requestsChange = (thisPeriod.requests + 1) / (lastPeriod.requests + 1) * 100 - 100;
		let successChange = ((thisPeriod.success + 1) / (thisPeriod.requests + 1) + 1) / ((lastPeriod.success + 1) / (lastPeriod.requests + 1) + 1) * 100 - 100;
		let responseTimeChange = ((thisPeriod.responseTime + 1) / (thisPeriod.requests + 1) + 1) / ((lastPeriod.responseTime + 1) / (lastPeriod.requests + 1) + 1) * 100 - 100;

		change = {
			requests: requestsChange,
			success: successChange,
			responseTime: responseTimeChange
		};
	}

	let change;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data, prevData } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.prevData === void 0 && $$bindings.prevData && prevData !== void 0) $$bindings.prevData(prevData);
	$$result.css.add(css$9);
	data && mounted && buildWeek();

	return `<div class="${"card svelte-1v5417n"}"><div class="${"card-title"}">Growth</div>
  <div class="${"values svelte-1v5417n"}">${change != undefined
	? `<div class="${"tile svelte-1v5417n"}"><div class="${"tile-value svelte-1v5417n"}"><span style="${"color: " + escape(change.requests > 0 ? 'var(--highlight)' : 'var(--red)', true)}">${escape(change.requests > 0 ? "+" : "")}${escape(change.requests.toFixed(1))}%</span></div>
        <div class="${"tile-label svelte-1v5417n"}">Requests</div></div>
      
      <div class="${"tile svelte-1v5417n"}"><div class="${"tile-value svelte-1v5417n"}"><span style="${"color: " + escape(change.success > 0 ? 'var(--highlight)' : 'var(--red)', true)}">${escape(change.success > 0 ? "+" : "")}${escape(change.success.toFixed(1))}%</span></div>
        <div class="${"tile-label svelte-1v5417n"}">Success rate</div></div>
      <div class="${"tile svelte-1v5417n"}"><div class="${"tile-value svelte-1v5417n"}"><span style="${"color: " + escape(
			change.responseTime < 0
			? 'var(--highlight)'
			: 'var(--red)',
			true
		)}">${escape(change.responseTime > 0 ? "+" : "")}${escape(change.responseTime.toFixed(1))}%</span></div>
        <div class="${"tile-label svelte-1v5417n"}">Response time</div></div>`
	: ``}</div>
</div>`;
});

/* src\components\dashboard\device\Browser.svelte generated by Svelte v3.53.1 */

const css$8 = {
	code: "#plotDiv.svelte-njiwdy{margin-right:20px}",
	map: "{\"version\":3,\"file\":\"Browser.svelte\",\"sources\":[\"Browser.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction getBrowser(userAgent) {\\r\\n    if (userAgent.match(/Seamonkey\\\\//)) {\\r\\n        return \\\"Seamonkey\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Firefox\\\\//)) {\\r\\n        return \\\"Firefox\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Chrome\\\\//)) {\\r\\n        return \\\"Chrome\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Chromium\\\\//)) {\\r\\n        return \\\"Chromium\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Safari\\\\//)) {\\r\\n        return \\\"Safari\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Edg\\\\//)) {\\r\\n        return \\\"Edge\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/OPR\\\\//) || userAgent.match(/Opera\\\\//)) {\\r\\n        return \\\"Opera\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/; MSIE /) || userAgent.match(/Trident\\\\//)) {\\r\\n        return \\\"Internet Explorer\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/curl\\\\//)) {\\r\\n        return \\\"Curl\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/PostmanRuntime\\\\//)) {\\r\\n        return \\\"Postman\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/insomnia\\\\//)) {\\r\\n        return \\\"Insomnia\\\";\\r\\n    }\\r\\n    else {\\r\\n        return \\\"Other\\\";\\r\\n    }\\r\\n}\\r\\nfunction browserPlotLayout() {\\r\\n    let monthAgo = new Date();\\r\\n    monthAgo.setDate(monthAgo.getDate() - 30);\\r\\n    let tomorrow = new Date();\\r\\n    tomorrow.setDate(tomorrow.getDate() + 1);\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 20, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 180,\\r\\n        width: 411,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Requests\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nlet colors = [\\r\\n    \\\"#3FCF8E\\\",\\r\\n    \\\"#5784BA\\\",\\r\\n    \\\"#EBEB81\\\",\\r\\n    \\\"#218B82\\\",\\r\\n    \\\"#FFD6A5\\\",\\r\\n    \\\"#F9968B\\\",\\r\\n    \\\"#B1A2CA\\\",\\r\\n    \\\"#E46161\\\", // Red\\r\\n];\\r\\nfunction pieChart() {\\r\\n    let browserCount = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let browser = getBrowser(data[i].user_agent);\\r\\n        if (!(browser in browserCount)) {\\r\\n            browserCount[browser] = 0;\\r\\n        }\\r\\n        browserCount[browser]++;\\r\\n    }\\r\\n    let browsers = [];\\r\\n    let count = [];\\r\\n    for (let browser in browserCount) {\\r\\n        browsers.push(browser);\\r\\n        count.push(browserCount[browser]);\\r\\n    }\\r\\n    return [\\r\\n        {\\r\\n            values: count,\\r\\n            labels: browsers,\\r\\n            type: \\\"pie\\\",\\r\\n            marker: {\\r\\n                colors: colors,\\r\\n            },\\r\\n        },\\r\\n    ];\\r\\n}\\r\\nfunction browserPlotData() {\\r\\n    return {\\r\\n        data: pieChart(),\\r\\n        layout: browserPlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genPlot() {\\r\\n    let plotData = browserPlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\nlet plotDiv;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && genPlot();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div id=\\\"plotly\\\">\\r\\n  <div id=\\\"plotDiv\\\" bind:this={plotDiv}>\\r\\n    <!-- Plotly chart will be drawn inside this DIV -->\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  #plotDiv {\\r\\n    margin-right: 20px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAqIE,QAAQ,cAAC,CAAC,AACR,YAAY,CAAE,IAAI,AACpB,CAAC\"}"
};

function getBrowser(userAgent) {
	if (userAgent.match(/Seamonkey\//)) {
		return "Seamonkey";
	} else if (userAgent.match(/Firefox\//)) {
		return "Firefox";
	} else if (userAgent.match(/Chrome\//)) {
		return "Chrome";
	} else if (userAgent.match(/Chromium\//)) {
		return "Chromium";
	} else if (userAgent.match(/Safari\//)) {
		return "Safari";
	} else if (userAgent.match(/Edg\//)) {
		return "Edge";
	} else if (userAgent.match(/OPR\//) || userAgent.match(/Opera\//)) {
		return "Opera";
	} else if (userAgent.match(/; MSIE /) || userAgent.match(/Trident\//)) {
		return "Internet Explorer";
	} else if (userAgent.match(/curl\//)) {
		return "Curl";
	} else if (userAgent.match(/PostmanRuntime\//)) {
		return "Postman";
	} else if (userAgent.match(/insomnia\//)) {
		return "Insomnia";
	} else {
		return "Other";
	}
}

function browserPlotLayout() {
	let monthAgo = new Date();
	monthAgo.setDate(monthAgo.getDate() - 30);
	let tomorrow = new Date();
	tomorrow.setDate(tomorrow.getDate() + 1);

	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 70, t: 20, b: 20, pad: 0 },
		hovermode: "closest",
		plot_bgcolor: "transparent",
		paper_bgcolor: "transparent",
		height: 180,
		width: 411,
		yaxis: {
			title: { text: "Requests" },
			gridcolor: "gray",
			showgrid: false,
			fixedrange: true
		},
		xaxis: { visible: false },
		dragmode: false
	};
}

const Browser = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let colors = [
		"#3FCF8E",
		"#5784BA",
		"#EBEB81",
		"#218B82",
		"#FFD6A5",
		"#F9968B",
		"#B1A2CA",
		"#E46161"
	]; // Red

	function pieChart() {
		let browserCount = {};

		for (let i = 0; i < data.length; i++) {
			let browser = getBrowser(data[i].user_agent);

			if (!(browser in browserCount)) {
				browserCount[browser] = 0;
			}

			browserCount[browser]++;
		}

		let browsers = [];
		let count = [];

		for (let browser in browserCount) {
			browsers.push(browser);
			count.push(browserCount[browser]);
		}

		return [
			{
				values: count,
				labels: browsers,
				type: "pie",
				marker: { colors }
			}
		];
	}

	function browserPlotData() {
		return {
			data: pieChart(),
			layout: browserPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot() {
		let plotData = browserPlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let plotDiv;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$8);
	data && mounted && genPlot();

	return `<div id="${"plotly"}"><div id="${"plotDiv"}" class="${"svelte-njiwdy"}"${add_attribute("this", plotDiv, 0)}></div>
</div>`;
});

/* src\components\dashboard\device\OperatingSystem.svelte generated by Svelte v3.53.1 */

const css$7 = {
	code: "#plotDiv.svelte-njiwdy{margin-right:20px}",
	map: "{\"version\":3,\"file\":\"OperatingSystem.svelte\",\"sources\":[\"OperatingSystem.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction getOS(userAgent) {\\r\\n    if (userAgent.match(/Win16/)) {\\r\\n        return \\\"Windows 3.11\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows 95)|(Win95)|(Windows_95)/)) {\\r\\n        return \\\"Windows 95\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows 98)|(Win98)/)) {\\r\\n        return \\\"Windows 98\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 5.0)|(Windows 2000)/)) {\\r\\n        return \\\"Windows 2000\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 5.1)|(Windows XP)/)) {\\r\\n        return \\\"Windows XP\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 5.2)/)) {\\r\\n        return \\\"Windows Server 2003\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 6.0)/)) {\\r\\n        return \\\"Windows Vista\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 6.1)/)) {\\r\\n        return \\\"Windows 7\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 6.2)/)) {\\r\\n        return \\\"Windows 8\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 10.0)/)) {\\r\\n        return \\\"Windows 10\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/)) {\\r\\n        return \\\"Windows NT 4.0\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Windows ME/)) {\\r\\n        return \\\"Windows ME\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/OpenBSD/)) {\\r\\n        return \\\"OpenBSE\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/SunOS/)) {\\r\\n        return \\\"SunOS\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Linux)|(X11)/)) {\\r\\n        return \\\"Linux\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(Mac_PowerPC)|(Macintosh)/)) {\\r\\n        return \\\"MacOS\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/QNX/)) {\\r\\n        return \\\"QNX\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/BeOS/)) {\\r\\n        return \\\"BeOS\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/OS\\\\/2/)) {\\r\\n        return \\\"OS/2\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/(nuhk)|(Googlebot)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\\\\/Teoma)|(ia_archiver)/)) {\\r\\n        return \\\"Search Bot\\\";\\r\\n    }\\r\\n    else {\\r\\n        return \\\"Unknown\\\";\\r\\n    }\\r\\n}\\r\\nfunction osPlotLayout() {\\r\\n    let monthAgo = new Date();\\r\\n    monthAgo.setDate(monthAgo.getDate() - 30);\\r\\n    let tomorrow = new Date();\\r\\n    tomorrow.setDate(tomorrow.getDate() + 1);\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 20, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 180,\\r\\n        width: 411,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Requests\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nlet colors = [\\r\\n    \\\"#3FCF8E\\\",\\r\\n    \\\"#5784BA\\\",\\r\\n    \\\"#EBEB81\\\",\\r\\n    \\\"#218B82\\\",\\r\\n    \\\"#FFD6A5\\\",\\r\\n    \\\"#F9968B\\\",\\r\\n    \\\"#B1A2CA\\\",\\r\\n    \\\"#E46161\\\", // Red\\r\\n];\\r\\nfunction pieChart() {\\r\\n    let osCount = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let os = getOS(data[i].user_agent);\\r\\n        if (!(os in osCount)) {\\r\\n            osCount[os] = 0;\\r\\n        }\\r\\n        osCount[os]++;\\r\\n    }\\r\\n    let os = [];\\r\\n    let count = [];\\r\\n    for (let browser in osCount) {\\r\\n        os.push(browser);\\r\\n        count.push(osCount[browser]);\\r\\n    }\\r\\n    return [\\r\\n        {\\r\\n            values: count,\\r\\n            labels: os,\\r\\n            type: \\\"pie\\\",\\r\\n            marker: {\\r\\n                colors: colors,\\r\\n            },\\r\\n        },\\r\\n    ];\\r\\n}\\r\\nfunction osPlotData() {\\r\\n    return {\\r\\n        data: pieChart(),\\r\\n        layout: osPlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genPlot() {\\r\\n    let plotData = osPlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\nlet plotDiv;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && genPlot();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div id=\\\"plotly\\\">\\r\\n  <div id=\\\"plotDiv\\\" bind:this={plotDiv}>\\r\\n    <!-- Plotly chart will be drawn inside this DIV -->\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  #plotDiv {\\r\\n    margin-right: 20px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAgKE,QAAQ,cAAC,CAAC,AACR,YAAY,CAAE,IAAI,AACpB,CAAC\"}"
};

function getOS(userAgent) {
	if (userAgent.match(/Win16/)) {
		return "Windows 3.11";
	} else if (userAgent.match(/(Windows 95)|(Win95)|(Windows_95)/)) {
		return "Windows 95";
	} else if (userAgent.match(/(Windows 98)|(Win98)/)) {
		return "Windows 98";
	} else if (userAgent.match(/(Windows NT 5.0)|(Windows 2000)/)) {
		return "Windows 2000";
	} else if (userAgent.match(/(Windows NT 5.1)|(Windows XP)/)) {
		return "Windows XP";
	} else if (userAgent.match(/(Windows NT 5.2)/)) {
		return "Windows Server 2003";
	} else if (userAgent.match(/(Windows NT 6.0)/)) {
		return "Windows Vista";
	} else if (userAgent.match(/(Windows NT 6.1)/)) {
		return "Windows 7";
	} else if (userAgent.match(/(Windows NT 6.2)/)) {
		return "Windows 8";
	} else if (userAgent.match(/(Windows NT 10.0)/)) {
		return "Windows 10";
	} else if (userAgent.match(/(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/)) {
		return "Windows NT 4.0";
	} else if (userAgent.match(/Windows ME/)) {
		return "Windows ME";
	} else if (userAgent.match(/OpenBSD/)) {
		return "OpenBSE";
	} else if (userAgent.match(/SunOS/)) {
		return "SunOS";
	} else if (userAgent.match(/(Linux)|(X11)/)) {
		return "Linux";
	} else if (userAgent.match(/(Mac_PowerPC)|(Macintosh)/)) {
		return "MacOS";
	} else if (userAgent.match(/QNX/)) {
		return "QNX";
	} else if (userAgent.match(/BeOS/)) {
		return "BeOS";
	} else if (userAgent.match(/OS\/2/)) {
		return "OS/2";
	} else if (userAgent.match(/(nuhk)|(Googlebot)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\/Teoma)|(ia_archiver)/)) {
		return "Search Bot";
	} else {
		return "Unknown";
	}
}

function osPlotLayout() {
	let monthAgo = new Date();
	monthAgo.setDate(monthAgo.getDate() - 30);
	let tomorrow = new Date();
	tomorrow.setDate(tomorrow.getDate() + 1);

	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 70, t: 20, b: 20, pad: 0 },
		hovermode: "closest",
		plot_bgcolor: "transparent",
		paper_bgcolor: "transparent",
		height: 180,
		width: 411,
		yaxis: {
			title: { text: "Requests" },
			gridcolor: "gray",
			showgrid: false,
			fixedrange: true
		},
		xaxis: { visible: false },
		dragmode: false
	};
}

const OperatingSystem = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let colors = [
		"#3FCF8E",
		"#5784BA",
		"#EBEB81",
		"#218B82",
		"#FFD6A5",
		"#F9968B",
		"#B1A2CA",
		"#E46161"
	]; // Red

	function pieChart() {
		let osCount = {};

		for (let i = 0; i < data.length; i++) {
			let os = getOS(data[i].user_agent);

			if (!(os in osCount)) {
				osCount[os] = 0;
			}

			osCount[os]++;
		}

		let os = [];
		let count = [];

		for (let browser in osCount) {
			os.push(browser);
			count.push(osCount[browser]);
		}

		return [
			{
				values: count,
				labels: os,
				type: "pie",
				marker: { colors }
			}
		];
	}

	function osPlotData() {
		return {
			data: pieChart(),
			layout: osPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot() {
		let plotData = osPlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let plotDiv;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$7);
	data && mounted && genPlot();

	return `<div id="${"plotly"}"><div id="${"plotDiv"}" class="${"svelte-njiwdy"}"${add_attribute("this", plotDiv, 0)}></div>
</div>`;
});

/* src\components\dashboard\device\DeviceType.svelte generated by Svelte v3.53.1 */

const css$6 = {
	code: "#plotDiv.svelte-njiwdy{margin-right:20px}",
	map: "{\"version\":3,\"file\":\"DeviceType.svelte\",\"sources\":[\"DeviceType.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction getDevice(userAgent) {\\r\\n    if (userAgent.match(/iPhone/)) {\\r\\n        return \\\"iPhone\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Android/)) {\\r\\n        return \\\"Android\\\";\\r\\n    }\\r\\n    else if (userAgent.match(/Tizen\\\\//)) {\\r\\n        return \\\"Samsung\\\";\\r\\n    }\\r\\n    else {\\r\\n        return \\\"Other\\\";\\r\\n    }\\r\\n}\\r\\nfunction devicePlotLayout() {\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 180,\\r\\n        width: 411,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Requests\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nlet colors = [\\r\\n    \\\"#3FCF8E\\\",\\r\\n    \\\"#E46161\\\",\\r\\n    \\\"#EBEB81\\\", // Yellow\\r\\n];\\r\\nfunction pieChart() {\\r\\n    let deviceCount = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let browser = getDevice(data[i].user_agent);\\r\\n        if (!(browser in deviceCount)) {\\r\\n            deviceCount[browser] = 0;\\r\\n        }\\r\\n        deviceCount[browser]++;\\r\\n    }\\r\\n    let device = [];\\r\\n    let count = [];\\r\\n    for (let browser in deviceCount) {\\r\\n        device.push(browser);\\r\\n        count.push(deviceCount[browser]);\\r\\n    }\\r\\n    return [\\r\\n        {\\r\\n            values: count,\\r\\n            labels: device,\\r\\n            type: \\\"pie\\\",\\r\\n            marker: {\\r\\n                colors: colors,\\r\\n            },\\r\\n        },\\r\\n    ];\\r\\n}\\r\\nfunction devicePlotData() {\\r\\n    return {\\r\\n        data: pieChart(),\\r\\n        layout: devicePlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genPlot() {\\r\\n    let plotData = devicePlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\nlet plotDiv;\\r\\nlet mounted = false;\\r\\nonMount(() => {\\r\\n    mounted = true;\\r\\n});\\r\\n$: data && mounted && genPlot();\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div id=\\\"plotly\\\">\\r\\n  <div id=\\\"plotDiv\\\" bind:this={plotDiv}>\\r\\n    <!-- Plotly chart will be drawn inside this DIV -->\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  #plotDiv {\\r\\n    margin-right: 20px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAoGE,QAAQ,cAAC,CAAC,AACR,YAAY,CAAE,IAAI,AACpB,CAAC\"}"
};

function getDevice(userAgent) {
	if (userAgent.match(/iPhone/)) {
		return "iPhone";
	} else if (userAgent.match(/Android/)) {
		return "Android";
	} else if (userAgent.match(/Tizen\//)) {
		return "Samsung";
	} else {
		return "Other";
	}
}

function devicePlotLayout() {
	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
		hovermode: "closest",
		plot_bgcolor: "transparent",
		paper_bgcolor: "transparent",
		height: 180,
		width: 411,
		yaxis: {
			title: { text: "Requests" },
			gridcolor: "gray",
			showgrid: false,
			fixedrange: true
		},
		xaxis: { visible: false },
		dragmode: false
	};
}

const DeviceType = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let colors = ["#3FCF8E", "#E46161", "#EBEB81"]; // Yellow

	function pieChart() {
		let deviceCount = {};

		for (let i = 0; i < data.length; i++) {
			let browser = getDevice(data[i].user_agent);

			if (!(browser in deviceCount)) {
				deviceCount[browser] = 0;
			}

			deviceCount[browser]++;
		}

		let device = [];
		let count = [];

		for (let browser in deviceCount) {
			device.push(browser);
			count.push(deviceCount[browser]);
		}

		return [
			{
				values: count,
				labels: device,
				type: "pie",
				marker: { colors }
			}
		];
	}

	function devicePlotData() {
		return {
			data: pieChart(),
			layout: devicePlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot() {
		let plotData = devicePlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let plotDiv;
	let mounted = false;

	onMount(() => {
		mounted = true;
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$6);
	data && mounted && genPlot();

	return `<div id="${"plotly"}"><div id="${"plotDiv"}" class="${"svelte-njiwdy"}"${add_attribute("this", plotDiv, 0)}></div>
</div>`;
});

/* src\components\dashboard\device\Device.svelte generated by Svelte v3.53.1 */

const css$5 = {
	code: ".card.svelte-sfsn3p{margin:2em 0 2em 1em;padding-bottom:1em;width:420px}.card-title.svelte-sfsn3p{display:flex}.toggle.svelte-sfsn3p{margin-left:auto}.active.svelte-sfsn3p{background:var(--highlight)}.os.svelte-sfsn3p,.browser.svelte-sfsn3p,.device.svelte-sfsn3p{display:none}button.svelte-sfsn3p{border:none;border-radius:4px;background:rgb(68, 68, 68);cursor:pointer;padding:2px 6px;margin-left:5px}@media screen and (max-width: 1580px){.card.svelte-sfsn3p{margin:0 0 2em;width:100%}}",
	map: "{\"version\":3,\"file\":\"Device.svelte\",\"sources\":[\"Device.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import Browser from \\\"./Browser.svelte\\\";\\r\\nimport OperatingSystem from \\\"./OperatingSystem.svelte\\\";\\r\\nimport DeviceType from \\\"./DeviceType.svelte\\\";\\r\\nfunction setBtn(target) {\\r\\n    activeBtn = target;\\r\\n    // Resize window to trigger new plot resize to match current card size\\r\\n    window.dispatchEvent(new Event(\\\"resize\\\"));\\r\\n}\\r\\nlet activeBtn = \\\"os\\\";\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Last week\\\">\\r\\n  <div class=\\\"card-title\\\">\\r\\n    Device\\r\\n\\r\\n    <div class=\\\"toggle\\\">\\r\\n      <button\\r\\n        class={activeBtn == \\\"os\\\" ? \\\"active\\\" : \\\"\\\"}\\r\\n        on:click={() => {\\r\\n          setBtn(\\\"os\\\");\\r\\n        }}>OS</button\\r\\n      >\\r\\n      <button\\r\\n        class={activeBtn == \\\"browser\\\" ? \\\"active\\\" : \\\"\\\"}\\r\\n        on:click={() => {\\r\\n          setBtn(\\\"browser\\\");\\r\\n        }}>Browser</button\\r\\n      >\\r\\n      <button\\r\\n        class={activeBtn == \\\"device\\\" ? \\\"active\\\" : \\\"\\\"}\\r\\n        on:click={() => {\\r\\n          setBtn(\\\"device\\\");\\r\\n        }}>Device</button\\r\\n      >\\r\\n    </div>\\r\\n  </div>\\r\\n  <div class=\\\"os\\\" style={activeBtn == \\\"os\\\" ? \\\"display: initial\\\" : \\\"\\\"}>\\r\\n    <OperatingSystem {data} />\\r\\n  </div>\\r\\n  <div class=\\\"browser\\\" style={activeBtn == \\\"browser\\\" ? \\\"display: initial\\\" : \\\"\\\"}>\\r\\n    <Browser {data} />\\r\\n  </div>\\r\\n  <div class=\\\"device\\\" style={activeBtn == \\\"device\\\" ? \\\"display: initial\\\" : \\\"\\\"}>\\r\\n    <DeviceType {data} />\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    margin: 2em 0 2em 1em;\\r\\n    padding-bottom: 1em;\\r\\n    width: 420px;\\r\\n  }\\r\\n  .card-title {\\r\\n    display: flex;\\r\\n  }\\r\\n  .toggle {\\r\\n    margin-left: auto;\\r\\n  }\\r\\n  .active {\\r\\n    background: var(--highlight);\\r\\n  }\\r\\n  .os,\\r\\n  .browser,\\r\\n  .device {\\r\\n    display: none;\\r\\n  }\\r\\n  button {\\r\\n    border: none;\\r\\n    border-radius: 4px;\\r\\n    background: rgb(68, 68, 68);\\r\\n    cursor: pointer;\\r\\n    padding: 2px 6px;\\r\\n    margin-left: 5px;\\r\\n  }\\r\\n  @media screen and (max-width: 1580px) {\\r\\n    .card {\\r\\n      margin: 0 0 2em;\\r\\n      width: 100%;\\r\\n    }\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAiDE,KAAK,cAAC,CAAC,AACL,MAAM,CAAE,GAAG,CAAC,CAAC,CAAC,GAAG,CAAC,GAAG,CACrB,cAAc,CAAE,GAAG,CACnB,KAAK,CAAE,KAAK,AACd,CAAC,AACD,WAAW,cAAC,CAAC,AACX,OAAO,CAAE,IAAI,AACf,CAAC,AACD,OAAO,cAAC,CAAC,AACP,WAAW,CAAE,IAAI,AACnB,CAAC,AACD,OAAO,cAAC,CAAC,AACP,UAAU,CAAE,IAAI,WAAW,CAAC,AAC9B,CAAC,AACD,iBAAG,CACH,sBAAQ,CACR,OAAO,cAAC,CAAC,AACP,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,cAAC,CAAC,AACN,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,UAAU,CAAE,IAAI,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,CAC3B,MAAM,CAAE,OAAO,CACf,OAAO,CAAE,GAAG,CAAC,GAAG,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,KAAK,cAAC,CAAC,AACL,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CACf,KAAK,CAAE,IAAI,AACb,CAAC,AACH,CAAC\"}"
};

const Device = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$5);

	return `<div class="${"card svelte-sfsn3p"}" title="${"Last week"}"><div class="${"card-title svelte-sfsn3p"}">Device

    <div class="${"toggle svelte-sfsn3p"}"><button class="${escape(null_to_empty("active" ), true) + " svelte-sfsn3p"}">OS</button>
      <button class="${escape(null_to_empty(""), true) + " svelte-sfsn3p"}">Browser</button>
      <button class="${escape(null_to_empty(""), true) + " svelte-sfsn3p"}">Device</button></div></div>
  <div class="${"os svelte-sfsn3p"}"${add_attribute("style", "display: initial" , 0)}>${validate_component(OperatingSystem, "OperatingSystem").$$render($$result, { data }, {}, {})}</div>
  <div class="${"browser svelte-sfsn3p"}"${add_attribute("style", "", 0)}>${validate_component(Browser, "Browser").$$render($$result, { data }, {}, {})}</div>
  <div class="${"device svelte-sfsn3p"}"${add_attribute("style", "", 0)}>${validate_component(DeviceType, "DeviceType").$$render($$result, { data }, {}, {})}</div>
</div>`;
});

/* src\routes\Dashboard.svelte generated by Svelte v3.53.1 */

const css$4 = {
	code: ".dashboard.svelte-1j1sa1i{min-height:90vh}.dashboard.svelte-1j1sa1i{margin:5em;display:flex;position:relative}.row.svelte-1j1sa1i{display:flex}.grid-row.svelte-1j1sa1i{display:flex}.right.svelte-1j1sa1i{flex-grow:1;margin-right:2em}.no-requests.svelte-1j1sa1i{height:70vh;font-size:1.5em;display:grid;place-items:center;color:var(--highlight)}.placeholder.svelte-1j1sa1i{min-height:85vh;display:grid;place-items:center}.time-period.svelte-1j1sa1i{position:absolute;display:flex;right:2em;top:-2.2em;border:1px solid #2e2e2e;border-radius:4px;overflow:hidden}.time-period-btn.svelte-1j1sa1i{background:#232323;padding:3px 12px;border:none;color:#707070;cursor:pointer}.time-period-btn-active.svelte-1j1sa1i{background:var(--highlight);color:black}@media screen and (max-width: 1580px){.grid-row.svelte-1j1sa1i{flex-direction:column}}@media screen and (max-width: 1100px){.dashboard.svelte-1j1sa1i{margin:2em 0}}",
	map: "{\"version\":3,\"file\":\"Dashboard.svelte\",\"sources\":[\"Dashboard.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nimport Requests from \\\"../components/dashboard/Requests.svelte\\\";\\r\\nimport Welcome from \\\"../components/dashboard/Welcome.svelte\\\";\\r\\nimport RequestsPerHour from \\\"../components/dashboard/RequestsPerHour.svelte\\\";\\r\\nimport ResponseTimes from \\\"../components/dashboard/ResponseTimes.svelte\\\";\\r\\nimport Endpoints from \\\"../components/dashboard/Endpoints.svelte\\\";\\r\\nimport Footer from \\\"../components/Footer.svelte\\\";\\r\\nimport SuccessRate from \\\"../components/dashboard/SuccessRate.svelte\\\";\\r\\nimport Activity from \\\"../components/dashboard/activity/Activity.svelte\\\";\\r\\nimport Version from \\\"../components/dashboard/Version.svelte\\\";\\r\\nimport UsageTime from \\\"../components/dashboard/UsageTime.svelte\\\";\\r\\nimport Growth from \\\"../components/dashboard/Growth.svelte\\\";\\r\\nimport Device from \\\"../components/dashboard/device/Device.svelte\\\";\\r\\nfunction formatUUID(userID) {\\r\\n    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(12, 16)}-${userID.slice(16, 20)}-${userID.slice(20)}`;\\r\\n}\\r\\nasync function fetchData() {\\r\\n    userID = formatUUID(userID);\\r\\n    // Fetch page ID\\r\\n    try {\\r\\n        const response = await fetch(`https://api-analytics-server.vercel.app/api/user-data/${userID}`);\\r\\n        if (response.status == 200) {\\r\\n            const json = await response.json();\\r\\n            data = json.value;\\r\\n            console.log(data);\\r\\n            setPeriod(\\\"month\\\");\\r\\n        }\\r\\n    }\\r\\n    catch (e) {\\r\\n        failed = true;\\r\\n    }\\r\\n}\\r\\nfunction inPeriod(date, days) {\\r\\n    let periodAgo = new Date();\\r\\n    periodAgo.setDate(periodAgo.getDate() - days);\\r\\n    return date > periodAgo;\\r\\n}\\r\\nfunction allTimePeriod(date) {\\r\\n    return true;\\r\\n}\\r\\nfunction periodToDays(period) {\\r\\n    if (period == \\\"24-hours\\\") {\\r\\n        return 1;\\r\\n    }\\r\\n    else if (period == \\\"week\\\") {\\r\\n        return 8;\\r\\n    }\\r\\n    else if (period == \\\"month\\\") {\\r\\n        return 30;\\r\\n    }\\r\\n    else if (period == \\\"3-months\\\") {\\r\\n        return 30 * 3;\\r\\n    }\\r\\n    else if (period == \\\"6-months\\\") {\\r\\n        return 30 * 6;\\r\\n    }\\r\\n    else if (period == \\\"year\\\") {\\r\\n        return 365;\\r\\n    }\\r\\n    else {\\r\\n        return null;\\r\\n    }\\r\\n}\\r\\nfunction setPeriodData() {\\r\\n    let days = periodToDays(period);\\r\\n    let counted = allTimePeriod;\\r\\n    if (days != null) {\\r\\n        counted = (date) => {\\r\\n            return inPeriod(date, days);\\r\\n        };\\r\\n    }\\r\\n    let dataSubset = [];\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        if (counted(date)) {\\r\\n            dataSubset.push(data[i]);\\r\\n        }\\r\\n    }\\r\\n    let count = 0;\\r\\n    for (let i = 0; i < dataSubset.length; i++) {\\r\\n        if (dataSubset[i].status >= 400 && dataSubset[i].status < 500) {\\r\\n            count++;\\r\\n        }\\r\\n    }\\r\\n    console.log(count);\\r\\n    periodData = dataSubset;\\r\\n}\\r\\nfunction inPrevPeriod(date, days) {\\r\\n    let startPeriodAgo = new Date();\\r\\n    startPeriodAgo.setDate(startPeriodAgo.getDate() - days * 2);\\r\\n    let endPeriodAgo = new Date();\\r\\n    endPeriodAgo.setDate(endPeriodAgo.getDate() - days);\\r\\n    return startPeriodAgo < date && date < endPeriodAgo;\\r\\n}\\r\\nfunction setPrevPeriodData() {\\r\\n    let days = periodToDays(period);\\r\\n    let inPeriod = allTimePeriod;\\r\\n    if (days != null) {\\r\\n        inPeriod = (date) => {\\r\\n            return inPrevPeriod(date, days);\\r\\n        };\\r\\n    }\\r\\n    let dataSubset = [];\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        if (inPeriod(date)) {\\r\\n            dataSubset.push(data[i]);\\r\\n        }\\r\\n    }\\r\\n    prevPeriodData = dataSubset;\\r\\n}\\r\\nfunction setPeriod(value) {\\r\\n    period = value;\\r\\n    setPeriodData();\\r\\n    setPrevPeriodData();\\r\\n}\\r\\nfunction getDemoStatus(date, status) {\\r\\n    let start = new Date();\\r\\n    start.setDate(start.getDate() - 100);\\r\\n    let end = new Date();\\r\\n    end.setDate(end.getDate() - 96);\\r\\n    if (date > start && date < end) {\\r\\n        return 400;\\r\\n    }\\r\\n    else {\\r\\n        return status;\\r\\n    }\\r\\n}\\r\\nfunction getDemoUserAgent() {\\r\\n    let p = Math.random();\\r\\n    if (p < 0.19) {\\r\\n        return \\\"Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1\\\";\\r\\n    }\\r\\n    else if (p < 0.3) {\\r\\n        return \\\"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0\\\";\\r\\n    }\\r\\n    else if (p < 0.34) {\\r\\n        return \\\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59\\\";\\r\\n    }\\r\\n    else if (p < 0.36) {\\r\\n        return \\\"curl/7.64.1\\\";\\r\\n    }\\r\\n    else if (p < 0.39) {\\r\\n        return \\\"PostmanRuntime/7.26.5\\\";\\r\\n    }\\r\\n    else if (p < 0.4) {\\r\\n        return \\\"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41\\\";\\r\\n    }\\r\\n    else if (p < 0.4) {\\r\\n        return \\\"Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0\\\";\\r\\n    }\\r\\n    else {\\r\\n        return \\\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36\\\";\\r\\n    }\\r\\n}\\r\\nfunction randomChoice(p) {\\r\\n    let rnd = p.reduce((a, b) => a + b) * Math.random();\\r\\n    return p.findIndex((a) => (rnd -= a) < 0);\\r\\n}\\r\\nfunction randomChoices(p, count) {\\r\\n    return Array.from(Array(count), randomChoice.bind(null, p));\\r\\n}\\r\\nfunction getHour() {\\r\\n    let p = Array(24).fill(1);\\r\\n    for (let i = 0; i < 3; i++) {\\r\\n        for (let j = 5 + i; j < 11 - i; j++) {\\r\\n            p[j] += 0.15;\\r\\n        }\\r\\n    }\\r\\n    for (let i = 0; i < 4; i++) {\\r\\n        for (let j = 11 + i; j < 21 - i; j++) {\\r\\n            p[j] += 0.15;\\r\\n        }\\r\\n    }\\r\\n    return randomChoices(p, 1)[0];\\r\\n}\\r\\nfunction addDemoSamples(demoData, endpoint, status, count, maxDaysAgo, minDaysAgo, maxResponseTime, minResponseTime) {\\r\\n    for (let i = 0; i < count; i++) {\\r\\n        let date = new Date();\\r\\n        date.setDate(date.getDate() - Math.floor(Math.random() * maxDaysAgo + minDaysAgo));\\r\\n        date.setHours(getHour());\\r\\n        demoData.push({\\r\\n            hostname: \\\"demo-api.com\\\",\\r\\n            path: endpoint,\\r\\n            user_agent: getDemoUserAgent(),\\r\\n            method: 0,\\r\\n            status: getDemoStatus(date, status),\\r\\n            response_time: Math.floor(Math.random() * maxResponseTime + minResponseTime),\\r\\n            created_at: date.toISOString(),\\r\\n        });\\r\\n    }\\r\\n}\\r\\nfunction genDemoData() {\\r\\n    let demoData = [];\\r\\n    addDemoSamples(demoData, \\\"/v1/\\\", 200, 18000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v1/\\\", 400, 1000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v1/account\\\", 200, 8000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v1/account\\\", 400, 1200, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v1/help\\\", 200, 700, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v1/help\\\", 400, 70, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 200, 35000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 400, 200, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/account\\\", 200, 14000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/account\\\", 400, 3000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/account/update\\\", 200, 6000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/account/update\\\", 400, 400, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/help\\\", 200, 6000, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/help\\\", 400, 400, 650, 0, 240, 55);\\r\\n    addDemoSamples(demoData, \\\"/v2/account\\\", 200, 16000, 450, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/account\\\", 400, 2000, 450, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/help\\\", 200, 8000, 300, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/help\\\", 400, 800, 300, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 200, 4000, 200, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 400, 800, 200, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 200, 3000, 100, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 400, 500, 100, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 200, 1000, 60, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 400, 50, 60, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 200, 500, 40, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 400, 25, 40, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 200, 250, 10, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 400, 20, 10, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 200, 125, 5, 0, 100, 30);\\r\\n    addDemoSamples(demoData, \\\"/v2/\\\", 400, 10, 5, 0, 100, 30);\\r\\n    data = demoData;\\r\\n    setPeriod(\\\"month\\\");\\r\\n}\\r\\nlet data;\\r\\nlet periodData;\\r\\nlet prevPeriodData;\\r\\nlet period = \\\"month\\\";\\r\\nlet failed = false;\\r\\nonMount(() => {\\r\\n    if (demo) {\\r\\n        genDemoData();\\r\\n    }\\r\\n    else {\\r\\n        fetchData();\\r\\n    }\\r\\n});\\r\\nexport let userID, demo;\\r\\n</script>\\r\\n\\r\\n{#if periodData != undefined}\\r\\n  <div class=\\\"dashboard\\\">\\r\\n    <div class=\\\"time-period\\\">\\r\\n      <button\\r\\n        class=\\\"time-period-btn {period == '24-hours'\\r\\n          ? 'time-period-btn-active'\\r\\n          : ''}\\\"\\r\\n        on:click={() => {\\r\\n          setPeriod(\\\"24-hours\\\");\\r\\n        }}\\r\\n      >\\r\\n        24 hours\\r\\n      </button>\\r\\n      <button\\r\\n        class=\\\"time-period-btn {period == 'week'\\r\\n          ? 'time-period-btn-active'\\r\\n          : ''}\\\"\\r\\n        on:click={() => {\\r\\n          setPeriod(\\\"week\\\");\\r\\n        }}\\r\\n      >\\r\\n        Week\\r\\n      </button>\\r\\n      <button\\r\\n        class=\\\"time-period-btn {period == 'month'\\r\\n          ? 'time-period-btn-active'\\r\\n          : ''}\\\"\\r\\n        on:click={() => {\\r\\n          setPeriod(\\\"month\\\");\\r\\n        }}\\r\\n      >\\r\\n        Month\\r\\n      </button>\\r\\n      <button\\r\\n        class=\\\"time-period-btn {period == '3-months'\\r\\n          ? 'time-period-btn-active'\\r\\n          : ''}\\\"\\r\\n        on:click={() => {\\r\\n          setPeriod(\\\"3-months\\\");\\r\\n        }}\\r\\n      >\\r\\n        3 months\\r\\n      </button>\\r\\n      <button\\r\\n        class=\\\"time-period-btn {period == '6-months'\\r\\n          ? 'time-period-btn-active'\\r\\n          : ''}\\\"\\r\\n        on:click={() => {\\r\\n          setPeriod(\\\"6-months\\\");\\r\\n        }}\\r\\n      >\\r\\n        6 months\\r\\n      </button>\\r\\n      <button\\r\\n        class=\\\"time-period-btn {period == 'year'\\r\\n          ? 'time-period-btn-active'\\r\\n          : ''}\\\"\\r\\n        on:click={() => {\\r\\n          setPeriod(\\\"year\\\");\\r\\n        }}\\r\\n      >\\r\\n        Year\\r\\n      </button>\\r\\n      <button\\r\\n        class=\\\"time-period-btn {period == 'all-time'\\r\\n          ? 'time-period-btn-active'\\r\\n          : ''}\\\"\\r\\n        on:click={() => {\\r\\n          setPeriod(\\\"all-time\\\");\\r\\n        }}\\r\\n      >\\r\\n        All time\\r\\n      </button>\\r\\n    </div>\\r\\n    <div class=\\\"left\\\">\\r\\n      <div class=\\\"row\\\">\\r\\n        <Welcome />\\r\\n        <SuccessRate data={periodData} />\\r\\n      </div>\\r\\n      <div class=\\\"row\\\">\\r\\n        <Requests data={periodData} prevData={prevPeriodData} />\\r\\n        <RequestsPerHour data={periodData} {period} />\\r\\n      </div>\\r\\n      <ResponseTimes data={periodData} />\\r\\n      <Endpoints data={periodData} />\\r\\n      <Version data={periodData} />\\r\\n    </div>\\r\\n    <div class=\\\"right\\\">\\r\\n      <Activity data={periodData} {period} />\\r\\n      <div class=\\\"grid-row\\\">\\r\\n        <Growth data={periodData} prevData={prevPeriodData} />\\r\\n        <Device data={periodData} />\\r\\n      </div>\\r\\n      <UsageTime data={periodData} />\\r\\n    </div>\\r\\n  </div>\\r\\n{:else if failed}\\r\\n  <div class=\\\"no-requests\\\">No requests currently logged.</div>\\r\\n{:else}\\r\\n  <div class=\\\"placeholder\\\" style=\\\"min-height: 85vh;\\\">\\r\\n    <div class=\\\"spinner\\\">\\r\\n      <div class=\\\"loader\\\" />\\r\\n    </div>\\r\\n  </div>\\r\\n{/if}\\r\\n<Footer />\\r\\n\\r\\n<style>\\r\\n  .dashboard {\\r\\n    min-height: 90vh;\\r\\n  }\\r\\n  .dashboard {\\r\\n    margin: 5em;\\r\\n    display: flex;\\r\\n    position: relative;\\r\\n  }\\r\\n  .row {\\r\\n    display: flex;\\r\\n  }\\r\\n  .grid-row {\\r\\n    display: flex;\\r\\n  }\\r\\n  .right {\\r\\n    flex-grow: 1;\\r\\n    margin-right: 2em;\\r\\n  }\\r\\n  .no-requests {\\r\\n    height: 70vh;\\r\\n    font-size: 1.5em;\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n    color: var(--highlight);\\r\\n  }\\r\\n  .placeholder {\\r\\n    min-height: 85vh;\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n  }\\r\\n  .time-period {\\r\\n    position: absolute;\\r\\n    display: flex;\\r\\n    right: 2em;\\r\\n    top: -2.2em;\\r\\n    border: 1px solid #2e2e2e;\\r\\n    border-radius: 4px;\\r\\n    overflow: hidden;\\r\\n  }\\r\\n  .time-period-btn {\\r\\n    background: #232323;\\r\\n    padding: 3px 12px;\\r\\n    border: none;\\r\\n    color: #707070;\\r\\n    cursor: pointer;\\r\\n  }\\r\\n  .time-period-btn-active {\\r\\n    background: var(--highlight);\\r\\n    color: black;\\r\\n  }\\r\\n  @media screen and (max-width: 1580px) {\\r\\n    .grid-row {\\r\\n      flex-direction: column;\\r\\n    }\\r\\n  }\\r\\n  @media screen and (max-width: 1100px) {\\r\\n    .dashboard {\\r\\n      margin: 2em 0;\\r\\n    }\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AA+VE,UAAU,eAAC,CAAC,AACV,UAAU,CAAE,IAAI,AAClB,CAAC,AACD,UAAU,eAAC,CAAC,AACV,MAAM,CAAE,GAAG,CACX,OAAO,CAAE,IAAI,CACb,QAAQ,CAAE,QAAQ,AACpB,CAAC,AACD,IAAI,eAAC,CAAC,AACJ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,SAAS,eAAC,CAAC,AACT,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,eAAC,CAAC,AACN,SAAS,CAAE,CAAC,CACZ,YAAY,CAAE,GAAG,AACnB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,MAAM,CAAE,IAAI,CACZ,SAAS,CAAE,KAAK,CAChB,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,CACnB,KAAK,CAAE,IAAI,WAAW,CAAC,AACzB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,UAAU,CAAE,IAAI,CAChB,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,QAAQ,CAAE,QAAQ,CAClB,OAAO,CAAE,IAAI,CACb,KAAK,CAAE,GAAG,CACV,GAAG,CAAE,MAAM,CACX,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,CACzB,aAAa,CAAE,GAAG,CAClB,QAAQ,CAAE,MAAM,AAClB,CAAC,AACD,gBAAgB,eAAC,CAAC,AAChB,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,GAAG,CAAC,IAAI,CACjB,MAAM,CAAE,IAAI,CACZ,KAAK,CAAE,OAAO,CACd,MAAM,CAAE,OAAO,AACjB,CAAC,AACD,uBAAuB,eAAC,CAAC,AACvB,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,KAAK,CAAE,KAAK,AACd,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,SAAS,eAAC,CAAC,AACT,cAAc,CAAE,MAAM,AACxB,CAAC,AACH,CAAC,AACD,OAAO,MAAM,CAAC,GAAG,CAAC,YAAY,MAAM,CAAC,AAAC,CAAC,AACrC,UAAU,eAAC,CAAC,AACV,MAAM,CAAE,GAAG,CAAC,CAAC,AACf,CAAC,AACH,CAAC\"}"
};

function formatUUID$1(userID) {
	return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(12, 16)}-${userID.slice(16, 20)}-${userID.slice(20)}`;
}

function inPeriod(date, days) {
	let periodAgo = new Date();
	periodAgo.setDate(periodAgo.getDate() - days);
	return date > periodAgo;
}

function allTimePeriod(date) {
	return true;
}

function periodToDays(period) {
	if (period == "24-hours") {
		return 1;
	} else if (period == "week") {
		return 8;
	} else if (period == "month") {
		return 30;
	} else if (period == "3-months") {
		return 30 * 3;
	} else if (period == "6-months") {
		return 30 * 6;
	} else if (period == "year") {
		return 365;
	} else {
		return null;
	}
}

function inPrevPeriod(date, days) {
	let startPeriodAgo = new Date();
	startPeriodAgo.setDate(startPeriodAgo.getDate() - days * 2);
	let endPeriodAgo = new Date();
	endPeriodAgo.setDate(endPeriodAgo.getDate() - days);
	return startPeriodAgo < date && date < endPeriodAgo;
}

function getDemoStatus(date, status) {
	let start = new Date();
	start.setDate(start.getDate() - 100);
	let end = new Date();
	end.setDate(end.getDate() - 96);

	if (date > start && date < end) {
		return 400;
	} else {
		return status;
	}
}

function getDemoUserAgent() {
	let p = Math.random();

	if (p < 0.19) {
		return "Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1";
	} else if (p < 0.3) {
		return "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0";
	} else if (p < 0.34) {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59";
	} else if (p < 0.36) {
		return "curl/7.64.1";
	} else if (p < 0.39) {
		return "PostmanRuntime/7.26.5";
	} else if (p < 0.4) {
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41";
	} else if (p < 0.4) {
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0";
	} else {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36";
	}
}

function randomChoice(p) {
	let rnd = p.reduce((a, b) => a + b) * Math.random();
	return p.findIndex(a => (rnd -= a) < 0);
}

function randomChoices(p, count) {
	return Array.from(Array(count), randomChoice.bind(null, p));
}

function getHour() {
	let p = Array(24).fill(1);

	for (let i = 0; i < 3; i++) {
		for (let j = 5 + i; j < 11 - i; j++) {
			p[j] += 0.15;
		}
	}

	for (let i = 0; i < 4; i++) {
		for (let j = 11 + i; j < 21 - i; j++) {
			p[j] += 0.15;
		}
	}

	return randomChoices(p, 1)[0];
}

function addDemoSamples(
	demoData,
endpoint,
status,
count,
maxDaysAgo,
minDaysAgo,
maxResponseTime,
minResponseTime
) {
	for (let i = 0; i < count; i++) {
		let date = new Date();
		date.setDate(date.getDate() - Math.floor(Math.random() * maxDaysAgo + minDaysAgo));
		date.setHours(getHour());

		demoData.push({
			hostname: "demo-api.com",
			path: endpoint,
			user_agent: getDemoUserAgent(),
			method: 0,
			status: getDemoStatus(date, status),
			response_time: Math.floor(Math.random() * maxResponseTime + minResponseTime),
			created_at: date.toISOString()
		});
	}
}

const Dashboard = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	async function fetchData() {
		userID = formatUUID$1(userID);

		// Fetch page ID
		try {
			const response = await fetch(`https://api-analytics-server.vercel.app/api/user-data/${userID}`);

			if (response.status == 200) {
				const json = await response.json();
				data = json.value;
				console.log(data);
				setPeriod("month");
			}
		} catch(e) {
			failed = true;
		}
	}

	function setPeriodData() {
		let days = periodToDays(period);
		let counted = allTimePeriod;

		if (days != null) {
			counted = date => {
				return inPeriod(date, days);
			};
		}

		let dataSubset = [];

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);

			if (counted(date)) {
				dataSubset.push(data[i]);
			}
		}

		let count = 0;

		for (let i = 0; i < dataSubset.length; i++) {
			if (dataSubset[i].status >= 400 && dataSubset[i].status < 500) {
				count++;
			}
		}

		console.log(count);
		periodData = dataSubset;
	}

	function setPrevPeriodData() {
		let days = periodToDays(period);
		let inPeriod = allTimePeriod;

		if (days != null) {
			inPeriod = date => {
				return inPrevPeriod(date, days);
			};
		}

		let dataSubset = [];

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);

			if (inPeriod(date)) {
				dataSubset.push(data[i]);
			}
		}

		prevPeriodData = dataSubset;
	}

	function setPeriod(value) {
		period = value;
		setPeriodData();
		setPrevPeriodData();
	}

	function genDemoData() {
		let demoData = [];
		addDemoSamples(demoData, "/v1/", 200, 18000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v1/", 400, 1000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v1/account", 200, 8000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v1/account", 400, 1200, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v1/help", 200, 700, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v1/help", 400, 70, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/", 200, 35000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/", 400, 200, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/account", 200, 14000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/account", 400, 3000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/account/update", 200, 6000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/account/update", 400, 400, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/help", 200, 6000, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/help", 400, 400, 650, 0, 240, 55);
		addDemoSamples(demoData, "/v2/account", 200, 16000, 450, 0, 100, 30);
		addDemoSamples(demoData, "/v2/account", 400, 2000, 450, 0, 100, 30);
		addDemoSamples(demoData, "/v2/help", 200, 8000, 300, 0, 100, 30);
		addDemoSamples(demoData, "/v2/help", 400, 800, 300, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 200, 4000, 200, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 400, 800, 200, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 200, 3000, 100, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 400, 500, 100, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 200, 1000, 60, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 400, 50, 60, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 200, 500, 40, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 400, 25, 40, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 200, 250, 10, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 400, 20, 10, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 200, 125, 5, 0, 100, 30);
		addDemoSamples(demoData, "/v2/", 400, 10, 5, 0, 100, 30);
		data = demoData;
		setPeriod("month");
	}

	let data;
	let periodData;
	let prevPeriodData;
	let period = "month";
	let failed = false;

	onMount(() => {
		if (demo) {
			genDemoData();
		} else {
			fetchData();
		}
	});

	let { userID, demo } = $$props;
	if ($$props.userID === void 0 && $$bindings.userID && userID !== void 0) $$bindings.userID(userID);
	if ($$props.demo === void 0 && $$bindings.demo && demo !== void 0) $$bindings.demo(demo);
	$$result.css.add(css$4);

	return `${periodData != undefined
	? `<div class="${"dashboard svelte-1j1sa1i"}"><div class="${"time-period svelte-1j1sa1i"}"><button class="${"time-period-btn " + escape(period == '24-hours' ? 'time-period-btn-active' : '', true) + " svelte-1j1sa1i"}">24 hours
      </button>
      <button class="${"time-period-btn " + escape(period == 'week' ? 'time-period-btn-active' : '', true) + " svelte-1j1sa1i"}">Week
      </button>
      <button class="${"time-period-btn " + escape(period == 'month' ? 'time-period-btn-active' : '', true) + " svelte-1j1sa1i"}">Month
      </button>
      <button class="${"time-period-btn " + escape(period == '3-months' ? 'time-period-btn-active' : '', true) + " svelte-1j1sa1i"}">3 months
      </button>
      <button class="${"time-period-btn " + escape(period == '6-months' ? 'time-period-btn-active' : '', true) + " svelte-1j1sa1i"}">6 months
      </button>
      <button class="${"time-period-btn " + escape(period == 'year' ? 'time-period-btn-active' : '', true) + " svelte-1j1sa1i"}">Year
      </button>
      <button class="${"time-period-btn " + escape(period == 'all-time' ? 'time-period-btn-active' : '', true) + " svelte-1j1sa1i"}">All time
      </button></div>
    <div class="${"left"}"><div class="${"row svelte-1j1sa1i"}">${validate_component(Welcome, "Welcome").$$render($$result, {}, {}, {})}
        ${validate_component(SuccessRate, "SuccessRate").$$render($$result, { data: periodData }, {}, {})}</div>
      <div class="${"row svelte-1j1sa1i"}">${validate_component(Requests, "Requests").$$render(
			$$result,
			{
				data: periodData,
				prevData: prevPeriodData
			},
			{},
			{}
		)}
        ${validate_component(RequestsPerHour, "RequestsPerHour").$$render($$result, { data: periodData, period }, {}, {})}</div>
      ${validate_component(ResponseTimes, "ResponseTimes").$$render($$result, { data: periodData }, {}, {})}
      ${validate_component(Endpoints, "Endpoints").$$render($$result, { data: periodData }, {}, {})}
      ${validate_component(Version, "Version").$$render($$result, { data: periodData }, {}, {})}</div>
    <div class="${"right svelte-1j1sa1i"}">${validate_component(Activity, "Activity").$$render($$result, { data: periodData, period }, {}, {})}
      <div class="${"grid-row svelte-1j1sa1i"}">${validate_component(Growth, "Growth").$$render(
			$$result,
			{
				data: periodData,
				prevData: prevPeriodData
			},
			{},
			{}
		)}
        ${validate_component(Device, "Device").$$render($$result, { data: periodData }, {}, {})}</div>
      ${validate_component(UsageTime, "UsageTime").$$render($$result, { data: periodData }, {}, {})}</div></div>`
	: `${failed
		? `<div class="${"no-requests svelte-1j1sa1i"}">No requests currently logged.</div>`
		: `<div class="${"placeholder svelte-1j1sa1i"}" style="${"min-height: 85vh;"}"><div class="${"spinner"}"><div class="${"loader"}"></div></div></div>`}`}
${validate_component(Footer, "Footer").$$render($$result, {}, {}, {})}`;
});

/* src\routes\Delete.svelte generated by Svelte v3.53.1 */

const css$3 = {
	code: ".generate.svelte-brqh97{display:grid;place-items:center}h2.svelte-brqh97{margin:0 0 1em;font-size:2em}.content.svelte-brqh97{width:fit-content;background:#343434;background:#1c1c1c;padding:3.5em 4em 4em;border-radius:9px;margin:20vh 0 2vh;height:400px}input.svelte-brqh97{background:#1c1c1c;background:#343434;border:none;padding:0 20px;width:310px;font-size:1em;text-align:center;height:40px;border-radius:4px;margin-bottom:2.5em;color:white;display:grid}button.svelte-brqh97{height:40px;border-radius:4px;padding:0 20px;border:none;cursor:pointer;width:100px}.highlight.svelte-brqh97{color:#3fcf8e}.details.svelte-brqh97{font-size:0.8em;margin-top:calc(15px + 1em)}.notification.svelte-brqh97{color:#3fcf8e;margin-top:32px;height:16px}#generateBtn.svelte-brqh97{background:#3fcf8e}",
	map: "{\"version\":3,\"file\":\"Delete.svelte\",\"sources\":[\"Delete.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">let apiKey = \\\"\\\";\\r\\nlet loading = false;\\r\\nlet message = \\\"\\\";\\r\\nasync function genAPIKey() {\\r\\n    loading = true;\\r\\n    // Fetch page ID\\r\\n    const response = await fetch(`https://api-analytics-server.vercel.app/api/delete/${apiKey}`);\\r\\n    if (response.status == 200) {\\r\\n        message = \\\"Deleted successfully\\\";\\r\\n    }\\r\\n    else {\\r\\n        message = \\\"Error: API key invalid\\\";\\r\\n    }\\r\\n    loading = false;\\r\\n}\\r\\n</script>\\r\\n\\r\\n<div class=\\\"generate\\\">\\r\\n  <div class=\\\"content\\\">\\r\\n    <h2>Delete all stored data</h2>\\r\\n    <input type=\\\"text\\\" bind:value={apiKey} placeholder=\\\"Enter API key\\\" />\\r\\n    <button id=\\\"generateBtn\\\" on:click={genAPIKey}>Delete</button>\\r\\n    <div class=\\\"notification\\\">{message}</div>\\r\\n    <div class=\\\"spinner\\\">\\r\\n      <div class=\\\"loader\\\" style=\\\"display: {loading ? 'initial' : 'none'}\\\" />\\r\\n    </div>\\r\\n  </div>\\r\\n  <div class=\\\"details\\\">\\r\\n    <!-- <div class=\\\"keep-secure\\\">Keep your API key safe and secure.</div> -->\\r\\n    <div class=\\\"highlight logo\\\">API Analytics</div>\\r\\n    <img class=\\\"footer-logo\\\" src=\\\"img/logo.png\\\" alt=\\\"\\\" />\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .generate {\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n  }\\r\\n  h2 {\\r\\n    margin: 0 0 1em;\\r\\n    font-size: 2em;\\r\\n  }\\r\\n  .content {\\r\\n    width: fit-content;\\r\\n    background: #343434;\\r\\n    background: #1c1c1c;\\r\\n    padding: 3.5em 4em 4em;\\r\\n    border-radius: 9px;\\r\\n    margin: 20vh 0 2vh;\\r\\n    height: 400px;\\r\\n  }\\r\\n  input {\\r\\n    background: #1c1c1c;\\r\\n    background: #343434;\\r\\n    border: none;\\r\\n    padding: 0 20px;\\r\\n    width: 310px;\\r\\n    font-size: 1em;\\r\\n    text-align: center;\\r\\n    height: 40px;\\r\\n    border-radius: 4px;\\r\\n    margin-bottom: 2.5em;\\r\\n    color: white;\\r\\n    display: grid;\\r\\n  }\\r\\n  button {\\r\\n    height: 40px;\\r\\n    border-radius: 4px;\\r\\n    padding: 0 20px;\\r\\n    border: none;\\r\\n    cursor: pointer;\\r\\n    width: 100px;\\r\\n  }\\r\\n  .highlight {\\r\\n    color: #3fcf8e;\\r\\n  }\\r\\n  .details {\\r\\n    font-size: 0.8em;\\r\\n    margin-top: calc(15px + 1em);\\r\\n  }\\r\\n  .notification {\\r\\n    color: #3fcf8e;\\r\\n    margin-top: 32px;\\r\\n    height: 16px;\\r\\n  }\\r\\n  #generateBtn {\\r\\n    background: #3fcf8e;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAmCE,SAAS,cAAC,CAAC,AACT,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,EAAE,cAAC,CAAC,AACF,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CACf,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,KAAK,CAAE,WAAW,CAClB,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,KAAK,CAAC,GAAG,CAAC,GAAG,CACtB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,CAAC,CAAC,CAAC,GAAG,CAClB,MAAM,CAAE,KAAK,AACf,CAAC,AACD,KAAK,cAAC,CAAC,AACL,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,MAAM,CAAE,IAAI,CACZ,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,KAAK,CAAE,KAAK,CACZ,SAAS,CAAE,GAAG,CACd,UAAU,CAAE,MAAM,CAClB,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,aAAa,CAAE,KAAK,CACpB,KAAK,CAAE,KAAK,CACZ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,cAAC,CAAC,AACN,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,OAAO,CACf,KAAK,CAAE,KAAK,AACd,CAAC,AACD,UAAU,cAAC,CAAC,AACV,KAAK,CAAE,OAAO,AAChB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,SAAS,CAAE,KAAK,CAChB,UAAU,CAAE,KAAK,IAAI,CAAC,CAAC,CAAC,GAAG,CAAC,AAC9B,CAAC,AACD,aAAa,cAAC,CAAC,AACb,KAAK,CAAE,OAAO,CACd,UAAU,CAAE,IAAI,CAChB,MAAM,CAAE,IAAI,AACd,CAAC,AACD,YAAY,cAAC,CAAC,AACZ,UAAU,CAAE,OAAO,AACrB,CAAC\"}"
};

const Delete = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let apiKey = "";
	let message = "";

	$$result.css.add(css$3);

	return `<div class="${"generate svelte-brqh97"}"><div class="${"content svelte-brqh97"}"><h2 class="${"svelte-brqh97"}">Delete all stored data</h2>
    <input type="${"text"}" placeholder="${"Enter API key"}" class="${"svelte-brqh97"}"${add_attribute("value", apiKey, 0)}>
    <button id="${"generateBtn"}" class="${"svelte-brqh97"}">Delete</button>
    <div class="${"notification svelte-brqh97"}">${escape(message)}</div>
    <div class="${"spinner"}"><div class="${"loader"}" style="${"display: " + escape('none', true)}"></div></div></div>
  <div class="${"details svelte-brqh97"}">
    <div class="${"highlight logo svelte-brqh97"}">API Analytics</div>
    <img class="${"footer-logo"}" src="${"img/logo.png"}" alt="${""}"></div>
</div>`;
});

/* src\components\monitoring\ResponseTime.svelte generated by Svelte v3.53.1 */

function periodToMarkers$1(period) {
	if (period == "24h") {
		return 24 * 2;
	} else if (period == "7d") {
		return 12 * 7;
	} else if (period == "30d") {
		return 30 * 4;
	} else if (period == "60d") {
		return 60 * 2;
	} else {
		return null;
	}
}

function defaultLayout() {
	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 35, t: 0, b: 30, pad: 0 },
		hovermode: "closest",
		plot_bgcolor: "transparent",
		paper_bgcolor: "transparent",
		height: 120,
		yaxis: {
			title: null,
			gridcolor: "gray",
			showgrid: false,
			fixedrange: true
		},
		xaxis: {
			title: { text: "Date" },
			showgrid: false,
			fixedrange: true,
			visible: false
		},
		dragmode: false
	};
}

const ResponseTime = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function bars() {
		let markers = periodToMarkers$1(period);
		let dates = [];

		for (let i = 0; i < markers; i++) {
			dates.push(i);
		}

		let requests = [];

		for (let i = 0; i < markers; i++) {
			requests.push(data[i].response_time);
		}

		return [
			{
				x: dates,
				y: requests,
				type: "lines",
				marker: { color: "#707070" },
				fill: "tonexty",
				hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	function buildPlotData() {
		return {
			data: bars(),
			layout: defaultLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot() {
		let plotData = buildPlotData();

		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let plotDiv;
	let setup = false;

	onMount(() => {
		genPlot();
		setup = true;
	});

	let { data, period } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.period === void 0 && $$bindings.period && period !== void 0) $$bindings.period(period);
	period && setup && genPlot();
	return `<div id="${"plotly"}"><div id="${"plotDiv"}"${add_attribute("this", plotDiv, 0)}></div></div>`;
});

/* src\components\monitoring\Card.svelte generated by Svelte v3.53.1 */

const css$2 = {
	code: ".card.svelte-14ep2gi{width:min(100%, 1000px);border:1px solid #2e2e2e;margin:2.2em auto}.card-error.svelte-14ep2gi{box-shadow:rgba(228, 98, 98, 0.5) 0px 15px 110px 0px,\r\n      rgba(0, 0, 0, 0.4) 0px 30px 60px -30px;border:2px solid rgba(228, 98, 98, 1)}.card-text.svelte-14ep2gi{display:flex;margin:2em 2em 0;font-size:0.9em}.card-text-left.svelte-14ep2gi{flex-grow:1;display:flex}.endpoint.svelte-14ep2gi{margin-left:10px;letter-spacing:0.01em}.measurements.svelte-14ep2gi{display:flex;padding:1em 2em 2em}.measurement.svelte-14ep2gi{margin:0 0.1%;flex:1;height:3em;border-radius:1px;background:var(--highlight);background:rgb(40, 40, 40)}.success.svelte-14ep2gi{background:var(--highlight)}.delayed.svelte-14ep2gi{background:rgb(199, 229, 125)}.error.svelte-14ep2gi{background:rgb(228, 98, 98)}.null.svelte-14ep2gi{color:#707070}.uptime.svelte-14ep2gi{color:#707070}",
	map: "{\"version\":3,\"file\":\"Card.svelte\",\"sources\":[\"Card.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nimport ResponseTime from \\\"./ResponseTime.svelte\\\";\\r\\nfunction setUptime() {\\r\\n    let success = 0;\\r\\n    let total = 0;\\r\\n    for (let i = 0; i < measurements.length; i++) {\\r\\n        if (measurements[i].status == \\\"success\\\" ||\\r\\n            measurements[i].status == \\\"delay\\\") {\\r\\n            success++;\\r\\n        }\\r\\n        total++;\\r\\n    }\\r\\n    let per = (success / total) * 100;\\r\\n    if (per == 100) {\\r\\n        uptime = \\\"100\\\";\\r\\n    }\\r\\n    else {\\r\\n        uptime = per.toFixed(1);\\r\\n    }\\r\\n}\\r\\nfunction periodToMarkers(period) {\\r\\n    if (period == \\\"24h\\\") {\\r\\n        return 24 * 2;\\r\\n    }\\r\\n    else if (period == \\\"7d\\\") {\\r\\n        return 12 * 7;\\r\\n    }\\r\\n    else if (period == \\\"30d\\\") {\\r\\n        return 30 * 4;\\r\\n    }\\r\\n    else if (period == \\\"60d\\\") {\\r\\n        return 60 * 2;\\r\\n    }\\r\\n    else {\\r\\n        return null;\\r\\n    }\\r\\n}\\r\\nfunction setMeasurements() {\\r\\n    let markers = periodToMarkers(period);\\r\\n    measurements = Array(markers).fill({ status: null, response_time: 0 });\\r\\n    let start = markers - data.measurements.length;\\r\\n    for (let i = 0; i < data.measurements.length; i++) {\\r\\n        measurements[i + start] = data.measurements[i];\\r\\n    }\\r\\n}\\r\\nfunction setError() {\\r\\n    error = measurements[measurements.length - 1].status == \\\"error\\\";\\r\\n    anyError = anyError || error;\\r\\n}\\r\\nfunction build() {\\r\\n    setMeasurements();\\r\\n    setError();\\r\\n    setUptime();\\r\\n}\\r\\nlet uptime = \\\"\\\";\\r\\nlet error = false;\\r\\nlet measurements;\\r\\nonMount(() => {\\r\\n    build();\\r\\n});\\r\\n$: period && build();\\r\\nexport let data, period, anyError;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" class:card-error={error}>\\r\\n  <div class=\\\"card-text\\\">\\r\\n    <div class=\\\"card-text-left\\\">\\r\\n      <div class=\\\"card-status\\\">\\r\\n        {#if error}\\r\\n          <img src=\\\"/img/smallcross.png\\\" alt=\\\"\\\" />\\r\\n        {:else}\\r\\n          <img src=\\\"/img/smalltick.png\\\" alt=\\\"\\\" />\\r\\n        {/if}\\r\\n      </div>\\r\\n      <div class=\\\"endpoint\\\">{data.name}</div>\\r\\n    </div>\\r\\n    <div class=\\\"card-text-right\\\">\\r\\n      <div class=\\\"uptime\\\">Uptime: {uptime}%</div>\\r\\n    </div>\\r\\n  </div>\\r\\n  <div class=\\\"measurements\\\">\\r\\n    {#each measurements as measurement}\\r\\n      <div class=\\\"measurement {measurement.status}\\\" />\\r\\n    {/each}\\r\\n  </div>\\r\\n  <div class=\\\"response-time\\\">\\r\\n    <ResponseTime data={measurements} {period} />\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style scoped>\\r\\n  .card {\\r\\n    width: min(100%, 1000px);\\r\\n    border: 1px solid #2e2e2e;\\r\\n    margin: 2.2em auto;\\r\\n  }\\r\\n  .card-error {\\r\\n    box-shadow: rgba(228, 98, 98, 0.5) 0px 15px 110px 0px,\\r\\n      rgba(0, 0, 0, 0.4) 0px 30px 60px -30px;\\r\\n    border: 2px solid rgba(228, 98, 98, 1);\\r\\n  }\\r\\n  .card-text {\\r\\n    display: flex;\\r\\n    margin: 2em 2em 0;\\r\\n    font-size: 0.9em;\\r\\n  }\\r\\n  .card-text-left {\\r\\n    flex-grow: 1;\\r\\n    display: flex;\\r\\n  }\\r\\n  .endpoint {\\r\\n    margin-left: 10px;\\r\\n    letter-spacing: 0.01em;\\r\\n  }\\r\\n  .measurements {\\r\\n    display: flex;\\r\\n    padding: 1em 2em 2em;\\r\\n  }\\r\\n  .measurement {\\r\\n    margin: 0 0.1%;\\r\\n    flex: 1;\\r\\n    height: 3em;\\r\\n    border-radius: 1px;\\r\\n    background: var(--highlight);\\r\\n    background: rgb(40, 40, 40);\\r\\n  }\\r\\n  .success {\\r\\n    background: var(--highlight);\\r\\n  }\\r\\n  .delayed {\\r\\n    background: rgb(199, 229, 125);\\r\\n  }\\r\\n  .error {\\r\\n    background: rgb(228, 98, 98);\\r\\n  }\\r\\n  .null {\\r\\n    color: #707070;\\r\\n  }\\r\\n  .uptime {\\r\\n    color: #707070;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AA2FE,KAAK,eAAC,CAAC,AACL,KAAK,CAAE,IAAI,IAAI,CAAC,CAAC,MAAM,CAAC,CACxB,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,CACzB,MAAM,CAAE,KAAK,CAAC,IAAI,AACpB,CAAC,AACD,WAAW,eAAC,CAAC,AACX,UAAU,CAAE,KAAK,GAAG,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAAC,IAAI,CAAC,KAAK,CAAC,GAAG,CAAC;MACpD,KAAK,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAAC,IAAI,CAAC,IAAI,CAAC,KAAK,CACxC,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,KAAK,GAAG,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,CAAC,CAAC,AACxC,CAAC,AACD,UAAU,eAAC,CAAC,AACV,OAAO,CAAE,IAAI,CACb,MAAM,CAAE,GAAG,CAAC,GAAG,CAAC,CAAC,CACjB,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,eAAe,eAAC,CAAC,AACf,SAAS,CAAE,CAAC,CACZ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,SAAS,eAAC,CAAC,AACT,WAAW,CAAE,IAAI,CACjB,cAAc,CAAE,MAAM,AACxB,CAAC,AACD,aAAa,eAAC,CAAC,AACb,OAAO,CAAE,IAAI,CACb,OAAO,CAAE,GAAG,CAAC,GAAG,CAAC,GAAG,AACtB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,MAAM,CAAE,CAAC,CAAC,IAAI,CACd,IAAI,CAAE,CAAC,CACP,MAAM,CAAE,GAAG,CACX,aAAa,CAAE,GAAG,CAClB,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,UAAU,CAAE,IAAI,EAAE,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,AAC7B,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,UAAU,CAAE,IAAI,WAAW,CAAC,AAC9B,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,UAAU,CAAE,IAAI,GAAG,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAAC,AAChC,CAAC,AACD,MAAM,eAAC,CAAC,AACN,UAAU,CAAE,IAAI,GAAG,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,AAC9B,CAAC,AACD,KAAK,eAAC,CAAC,AACL,KAAK,CAAE,OAAO,AAChB,CAAC,AACD,OAAO,eAAC,CAAC,AACP,KAAK,CAAE,OAAO,AAChB,CAAC\"}"
};

function periodToMarkers(period) {
	if (period == "24h") {
		return 24 * 2;
	} else if (period == "7d") {
		return 12 * 7;
	} else if (period == "30d") {
		return 30 * 4;
	} else if (period == "60d") {
		return 60 * 2;
	} else {
		return null;
	}
}

const Card = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function setUptime() {
		let success = 0;
		let total = 0;

		for (let i = 0; i < measurements.length; i++) {
			if (measurements[i].status == "success" || measurements[i].status == "delay") {
				success++;
			}

			total++;
		}

		let per = success / total * 100;

		if (per == 100) {
			uptime = "100";
		} else {
			uptime = per.toFixed(1);
		}
	}

	function setMeasurements() {
		let markers = periodToMarkers(period);
		measurements = Array(markers).fill({ status: null, response_time: 0 });
		let start = markers - data.measurements.length;

		for (let i = 0; i < data.measurements.length; i++) {
			measurements[i + start] = data.measurements[i];
		}
	}

	function setError() {
		error = measurements[measurements.length - 1].status == "error";
		anyError = anyError || error;
	}

	function build() {
		setMeasurements();
		setError();
		setUptime();
	}

	let uptime = "";
	let error = false;
	let measurements;

	onMount(() => {
		build();
	});

	let { data, period, anyError } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	if ($$props.period === void 0 && $$bindings.period && period !== void 0) $$bindings.period(period);
	if ($$props.anyError === void 0 && $$bindings.anyError && anyError !== void 0) $$bindings.anyError(anyError);
	$$result.css.add(css$2);
	period && build();

	return `<div class="${["card svelte-14ep2gi", error ? "card-error" : ""].join(' ').trim()}"><div class="${"card-text svelte-14ep2gi"}"><div class="${"card-text-left svelte-14ep2gi"}"><div class="${"card-status"}">${error
	? `<img src="${"/img/smallcross.png"}" alt="${""}">`
	: `<img src="${"/img/smalltick.png"}" alt="${""}">`}</div>
      <div class="${"endpoint svelte-14ep2gi"}">${escape(data.name)}</div></div>
    <div class="${"card-text-right"}"><div class="${"uptime svelte-14ep2gi"}">Uptime: ${escape(uptime)}%</div></div></div>
  <div class="${"measurements svelte-14ep2gi"}">${each(measurements, measurement => {
		return `<div class="${"measurement " + escape(measurement.status, true) + " svelte-14ep2gi"}"></div>`;
	})}</div>
  <div class="${"response-time"}">${validate_component(ResponseTime, "ResponseTime").$$render($$result, { data: measurements, period }, {}, {})}</div>
</div>`;
});

/* src\components\monitoring\TrackNew.svelte generated by Svelte v3.53.1 */

const css$1 = {
	code: ".card.svelte-m15jzh{width:min(100%, 1000px);border:1px solid #2e2e2e;margin:2.2em auto}.card-text.svelte-m15jzh{margin:2em 2em 1.9em}input.svelte-m15jzh{background:#1c1c1c;border-radius:4px;border:none;margin:0 10px;width:100%;padding:0 5px;color:white}.url.svelte-m15jzh{display:flex}.url-text.svelte-m15jzh{margin:auto}.detail.svelte-m15jzh{margin-top:30px;color:#707070;font-weight:500;font-size:0.9em}button.svelte-m15jzh{background:var(--highlight);padding:5px 20px;border-radius:4px;border:none}",
	map: "{\"version\":3,\"file\":\"TrackNew.svelte\",\"sources\":[\"TrackNew.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\"></script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-text\\\">\\r\\n    <div class=\\\"url\\\">\\r\\n      <div class=\\\"url-text\\\">URL</div>\\r\\n      <input type=\\\"text\\\" />\\r\\n      <button>Add</button>\\r\\n    </div>\\r\\n    <div class=\\\"detail\\\">\\r\\n      Endpoints are pinged by our servers every 30 mins and response <b\\r\\n        >status</b\\r\\n      >\\r\\n      and response <b>time</b> are logged.\\r\\n    </div>\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style scoped>\\r\\n  .card {\\r\\n    width: min(100%, 1000px);\\r\\n    border: 1px solid #2e2e2e;\\r\\n    margin: 2.2em auto;\\r\\n  }\\r\\n  .card-text {\\r\\n    margin: 2em 2em 1.9em;\\r\\n  }\\r\\n  input {\\r\\n    background: #1c1c1c;\\r\\n    border-radius: 4px;\\r\\n    border: none;\\r\\n    margin: 0 10px;\\r\\n    width: 100%;\\r\\n    padding: 0 5px;\\r\\n    color: white;\\r\\n  }\\r\\n  .url {\\r\\n    display: flex;\\r\\n  }\\r\\n  .url-text {\\r\\n    margin: auto;\\r\\n  }\\r\\n  .detail {\\r\\n    margin-top: 30px;\\r\\n    color: #707070;\\r\\n    font-weight: 500;\\r\\n    font-size: 0.9em;\\r\\n  }\\r\\n  button {\\r\\n    background: var(--highlight);\\r\\n    padding: 5px 20px;\\r\\n    border-radius: 4px;\\r\\n    border: none;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAmBE,KAAK,cAAC,CAAC,AACL,KAAK,CAAE,IAAI,IAAI,CAAC,CAAC,MAAM,CAAC,CACxB,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,CACzB,MAAM,CAAE,KAAK,CAAC,IAAI,AACpB,CAAC,AACD,UAAU,cAAC,CAAC,AACV,MAAM,CAAE,GAAG,CAAC,GAAG,CAAC,KAAK,AACvB,CAAC,AACD,KAAK,cAAC,CAAC,AACL,UAAU,CAAE,OAAO,CACnB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,CAAC,CAAC,IAAI,CACd,KAAK,CAAE,IAAI,CACX,OAAO,CAAE,CAAC,CAAC,GAAG,CACd,KAAK,CAAE,KAAK,AACd,CAAC,AACD,IAAI,cAAC,CAAC,AACJ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,SAAS,cAAC,CAAC,AACT,MAAM,CAAE,IAAI,AACd,CAAC,AACD,OAAO,cAAC,CAAC,AACP,UAAU,CAAE,IAAI,CAChB,KAAK,CAAE,OAAO,CACd,WAAW,CAAE,GAAG,CAChB,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,OAAO,CAAE,GAAG,CAAC,IAAI,CACjB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,AACd,CAAC\"}"
};

const TrackNew = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	$$result.css.add(css$1);

	return `<div class="${"card svelte-m15jzh"}"><div class="${"card-text svelte-m15jzh"}"><div class="${"url svelte-m15jzh"}"><div class="${"url-text svelte-m15jzh"}">URL</div>
      <input type="${"text"}" class="${"svelte-m15jzh"}">
      <button class="${"svelte-m15jzh"}">Add</button></div>
    <div class="${"detail svelte-m15jzh"}">Endpoints are pinged by our servers every 30 mins and response <b>status</b>
      and response <b>time</b> are logged.
    </div></div>
</div>`;
});

/* src\routes\Monitoring.svelte generated by Svelte v3.53.1 */

const css = {
	code: ".monitoring.svelte-11gf7xd{font-weight:600}.status.svelte-11gf7xd{margin:13vh 0 9vh;display:grid;place-items:center}#status-image.svelte-11gf7xd{width:130px;margin-bottom:1em;filter:saturate(1.3)}.status-text.svelte-11gf7xd{font-size:2.2em;font-weight:700;color:white}.cards-container.svelte-11gf7xd{width:60%;margin:auto;padding-bottom:1em}.controls.svelte-11gf7xd{margin:auto;width:60%;width:min(100%, 1000px);display:flex}.add-new.svelte-11gf7xd{flex-grow:1;display:flex;justify-content:left}.period-controls.svelte-11gf7xd{margin-left:auto;display:flex;justify-content:right}.period-controls.svelte-11gf7xd{border:1px solid #2e2e2e;width:fit-content;border-radius:4px;overflow:hidden}button.svelte-11gf7xd{background:#232323;color:#707070;border:none;padding:3px 12px;cursor:pointer}.add-new-btn.svelte-11gf7xd{border:1px solid #2e2e2e;border-radius:4px}.add-new-text.svelte-11gf7xd{display:flex}.active.svelte-11gf7xd{background:var(--highlight);color:black}.plus.svelte-11gf7xd{padding-right:0.6em}",
	map: "{\"version\":3,\"file\":\"Monitoring.svelte\",\"sources\":[\"Monitoring.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nimport Footer from \\\"../components/Footer.svelte\\\";\\r\\nimport Card from \\\"../components/monitoring/Card.svelte\\\";\\r\\nimport TrackNew from \\\"../components/monitoring/TrackNew.svelte\\\";\\r\\nfunction formatUUID(userID) {\\r\\n    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(12, 16)}-${userID.slice(16, 20)}-${userID.slice(20)}`;\\r\\n}\\r\\nasync function fetchData() {\\r\\n    userID = formatUUID(userID);\\r\\n    // Fetch page ID\\r\\n    try {\\r\\n        const response = await fetch(`https://api-analytics-server.vercel.app/api/user-data/${userID}`);\\r\\n        if (response.status == 200) {\\r\\n            const json = await response.json();\\r\\n            data = json.value;\\r\\n            console.log(data);\\r\\n        }\\r\\n    }\\r\\n    catch (e) {\\r\\n        failed = true;\\r\\n    }\\r\\n}\\r\\nfunction setPeriod(value) {\\r\\n    period = value;\\r\\n    error = false;\\r\\n}\\r\\nfunction toggleShowTrackNew() {\\r\\n    showTrackNew = !showTrackNew;\\r\\n}\\r\\nlet error = false;\\r\\nlet period = \\\"30d\\\";\\r\\nlet data;\\r\\nlet measurements = Array(3);\\r\\nlet failed = false;\\r\\nfor (let i = 0; i < measurements.length; i++) {\\r\\n    measurements[i] = {\\r\\n        name: \\\"persona-api.vercel.app/v1/england\\\",\\r\\n        measurements: [],\\r\\n    };\\r\\n    for (let j = 0; j < 140; j++) {\\r\\n        measurements[i].measurements.push({\\r\\n            status: \\\"success\\\",\\r\\n            response_time: Math.random() * 10 + 5,\\r\\n        });\\r\\n    }\\r\\n}\\r\\nfor (let i = 50; i < 58; i++) {\\r\\n    measurements[0].measurements[i] = { status: \\\"error\\\", response_time: 0 };\\r\\n}\\r\\nmeasurements[1].name = \\\"persona-api.vercel.app/v1/england/features\\\";\\r\\nlet showTrackNew = false;\\r\\nonMount(() => {\\r\\n    fetchData();\\r\\n});\\r\\nexport let userID;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"monitoring\\\">\\r\\n  <div class=\\\"status\\\">\\r\\n    {#if error}\\r\\n      <div class=\\\"status-image\\\">\\r\\n        <img id=\\\"status-image\\\" src=\\\"/img/bigcross.png\\\" alt=\\\"\\\" />\\r\\n        <div class=\\\"status-text\\\">Systems down</div>\\r\\n      </div>\\r\\n    {:else}\\r\\n      <div class=\\\"status-image\\\">\\r\\n        <img id=\\\"status-image\\\" src=\\\"/img/bigtick.png\\\" alt=\\\"\\\" />\\r\\n        <div class=\\\"status-text\\\">Systems Online</div>\\r\\n      </div>\\r\\n    {/if}\\r\\n  </div>\\r\\n  <div class=\\\"cards-container\\\">\\r\\n    <div class=\\\"controls\\\">\\r\\n      <div class=\\\"add-new\\\">\\r\\n        <button class=\\\"add-new-btn\\\" on:click={toggleShowTrackNew}\\r\\n          ><div class=\\\"add-new-text\\\">\\r\\n            <span class=\\\"plus\\\">+</span> New\\r\\n          </div>\\r\\n        </button>\\r\\n      </div>\\r\\n      <div class=\\\"period-controls-container\\\">\\r\\n        <div class=\\\"period-controls\\\">\\r\\n          <button\\r\\n            class=\\\"period-btn {period == '24h' ? 'active' : ''}\\\"\\r\\n            on:click={() => {\\r\\n              setPeriod(\\\"24h\\\");\\r\\n            }}\\r\\n          >\\r\\n            24h\\r\\n          </button>\\r\\n          <button\\r\\n            class=\\\"period-btn {period == '7d' ? 'active' : ''}\\\"\\r\\n            on:click={() => {\\r\\n              setPeriod(\\\"7d\\\");\\r\\n            }}\\r\\n          >\\r\\n            7d\\r\\n          </button>\\r\\n          <button\\r\\n            class=\\\"period-btn {period == '30d' ? 'active' : ''}\\\"\\r\\n            on:click={() => {\\r\\n              setPeriod(\\\"30d\\\");\\r\\n            }}\\r\\n          >\\r\\n            30d\\r\\n          </button>\\r\\n          <button\\r\\n            class=\\\"period-btn {period == '60d' ? 'active' : ''}\\\"\\r\\n            on:click={() => {\\r\\n              setPeriod(\\\"60d\\\");\\r\\n            }}\\r\\n          >\\r\\n            60d\\r\\n          </button>\\r\\n        </div>\\r\\n      </div>\\r\\n    </div>\\r\\n    {#if showTrackNew || measurements.length == 0}\\r\\n      <TrackNew />\\r\\n    {/if}\\r\\n    <Card data={measurements[0]} {period} bind:anyError={error} />\\r\\n    <Card data={measurements[1]} {period} bind:anyError={error} />\\r\\n    <Card data={measurements[2]} {period} bind:anyError={error} />\\r\\n  </div>\\r\\n</div>\\r\\n<Footer />\\r\\n\\r\\n<style scoped>\\r\\n  .monitoring {\\r\\n    font-weight: 600;\\r\\n  }\\r\\n  .status {\\r\\n    margin: 13vh 0 9vh;\\r\\n    display: grid;\\r\\n    place-items: center;\\r\\n  }\\r\\n  #status-image {\\r\\n    width: 130px;\\r\\n    margin-bottom: 1em;\\r\\n    filter: saturate(1.3);\\r\\n  }\\r\\n  .status-text {\\r\\n    font-size: 2.2em;\\r\\n    font-weight: 700;\\r\\n    color: white;\\r\\n  }\\r\\n\\r\\n  .cards-container {\\r\\n    width: 60%;\\r\\n    margin: auto;\\r\\n    padding-bottom: 1em;\\r\\n  }\\r\\n\\r\\n  .controls {\\r\\n    margin: auto;\\r\\n    width: 60%;\\r\\n\\r\\n    width: min(100%, 1000px);\\r\\n    display: flex;\\r\\n  }\\r\\n  .add-new {\\r\\n    flex-grow: 1;\\r\\n    display: flex;\\r\\n    justify-content: left;\\r\\n  }\\r\\n  .period-controls {\\r\\n    margin-left: auto;\\r\\n    display: flex;\\r\\n    justify-content: right;\\r\\n  }\\r\\n\\r\\n  .period-controls {\\r\\n    border: 1px solid #2e2e2e;\\r\\n    width: fit-content;\\r\\n    border-radius: 4px;\\r\\n    overflow: hidden;\\r\\n  }\\r\\n\\r\\n  button {\\r\\n    background: #232323;\\r\\n    color: #707070;\\r\\n    border: none;\\r\\n    padding: 3px 12px;\\r\\n    cursor: pointer;\\r\\n  }\\r\\n  .add-new-btn {\\r\\n    border: 1px solid #2e2e2e;\\r\\n    border-radius: 4px;\\r\\n  }\\r\\n  .add-new-text {\\r\\n    display: flex;\\r\\n  }\\r\\n  .active {\\r\\n    background: var(--highlight);\\r\\n    color: black;\\r\\n  }\\r\\n  .plus {\\r\\n    padding-right: 0.6em;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAgIE,WAAW,eAAC,CAAC,AACX,WAAW,CAAE,GAAG,AAClB,CAAC,AACD,OAAO,eAAC,CAAC,AACP,MAAM,CAAE,IAAI,CAAC,CAAC,CAAC,GAAG,CAClB,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,aAAa,eAAC,CAAC,AACb,KAAK,CAAE,KAAK,CACZ,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,SAAS,GAAG,CAAC,AACvB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,CAChB,KAAK,CAAE,KAAK,AACd,CAAC,AAED,gBAAgB,eAAC,CAAC,AAChB,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,cAAc,CAAE,GAAG,AACrB,CAAC,AAED,SAAS,eAAC,CAAC,AACT,MAAM,CAAE,IAAI,CACZ,KAAK,CAAE,GAAG,CAEV,KAAK,CAAE,IAAI,IAAI,CAAC,CAAC,MAAM,CAAC,CACxB,OAAO,CAAE,IAAI,AACf,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,SAAS,CAAE,CAAC,CACZ,OAAO,CAAE,IAAI,CACb,eAAe,CAAE,IAAI,AACvB,CAAC,AACD,gBAAgB,eAAC,CAAC,AAChB,WAAW,CAAE,IAAI,CACjB,OAAO,CAAE,IAAI,CACb,eAAe,CAAE,KAAK,AACxB,CAAC,AAED,gBAAgB,eAAC,CAAC,AAChB,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,CACzB,KAAK,CAAE,WAAW,CAClB,aAAa,CAAE,GAAG,CAClB,QAAQ,CAAE,MAAM,AAClB,CAAC,AAED,MAAM,eAAC,CAAC,AACN,UAAU,CAAE,OAAO,CACnB,KAAK,CAAE,OAAO,CACd,MAAM,CAAE,IAAI,CACZ,OAAO,CAAE,GAAG,CAAC,IAAI,CACjB,MAAM,CAAE,OAAO,AACjB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,MAAM,CAAE,GAAG,CAAC,KAAK,CAAC,OAAO,CACzB,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,aAAa,eAAC,CAAC,AACb,OAAO,CAAE,IAAI,AACf,CAAC,AACD,OAAO,eAAC,CAAC,AACP,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,KAAK,CAAE,KAAK,AACd,CAAC,AACD,KAAK,eAAC,CAAC,AACL,aAAa,CAAE,KAAK,AACtB,CAAC\"}"
};

function formatUUID(userID) {
	return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(12, 16)}-${userID.slice(16, 20)}-${userID.slice(20)}`;
}

const Monitoring = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	async function fetchData() {
		userID = formatUUID(userID);

		// Fetch page ID
		try {
			const response = await fetch(`https://api-analytics-server.vercel.app/api/user-data/${userID}`);

			if (response.status == 200) {
				const json = await response.json();
				data = json.value;
				console.log(data);
			}
		} catch(e) {
		}
	}

	let error = false;
	let period = "30d";
	let data;
	let measurements = Array(3);

	for (let i = 0; i < measurements.length; i++) {
		measurements[i] = {
			name: "persona-api.vercel.app/v1/england",
			measurements: []
		};

		for (let j = 0; j < 140; j++) {
			measurements[i].measurements.push({
				status: "success",
				response_time: Math.random() * 10 + 5
			});
		}
	}

	for (let i = 50; i < 58; i++) {
		measurements[0].measurements[i] = { status: "error", response_time: 0 };
	}

	measurements[1].name = "persona-api.vercel.app/v1/england/features";

	onMount(() => {
		fetchData();
	});

	let { userID } = $$props;
	if ($$props.userID === void 0 && $$bindings.userID && userID !== void 0) $$bindings.userID(userID);
	$$result.css.add(css);
	let $$settled;
	let $$rendered;

	do {
		$$settled = true;

		$$rendered = `<div class="${"monitoring svelte-11gf7xd"}"><div class="${"status svelte-11gf7xd"}">${error
		? `<div class="${"status-image"}"><img id="${"status-image"}" src="${"/img/bigcross.png"}" alt="${""}" class="${"svelte-11gf7xd"}">
        <div class="${"status-text svelte-11gf7xd"}">Systems down</div></div>`
		: `<div class="${"status-image"}"><img id="${"status-image"}" src="${"/img/bigtick.png"}" alt="${""}" class="${"svelte-11gf7xd"}">
        <div class="${"status-text svelte-11gf7xd"}">Systems Online</div></div>`}</div>
  <div class="${"cards-container svelte-11gf7xd"}"><div class="${"controls svelte-11gf7xd"}"><div class="${"add-new svelte-11gf7xd"}"><button class="${"add-new-btn svelte-11gf7xd"}"><div class="${"add-new-text svelte-11gf7xd"}"><span class="${"plus svelte-11gf7xd"}">+</span> New
          </div></button></div>
      <div class="${"period-controls-container"}"><div class="${"period-controls svelte-11gf7xd"}"><button class="${"period-btn " + escape('', true) + " svelte-11gf7xd"}">24h
          </button>
          <button class="${"period-btn " + escape('', true) + " svelte-11gf7xd"}">7d
          </button>
          <button class="${"period-btn " + escape('active' , true) + " svelte-11gf7xd"}">30d
          </button>
          <button class="${"period-btn " + escape('', true) + " svelte-11gf7xd"}">60d
          </button></div></div></div>
    ${measurements.length == 0
		? `${validate_component(TrackNew, "TrackNew").$$render($$result, {}, {}, {})}`
		: ``}
    ${validate_component(Card, "Card").$$render(
			$$result,
			{
				data: measurements[0],
				period,
				anyError: error
			},
			{
				anyError: $$value => {
					error = $$value;
					$$settled = false;
				}
			},
			{}
		)}
    ${validate_component(Card, "Card").$$render(
			$$result,
			{
				data: measurements[1],
				period,
				anyError: error
			},
			{
				anyError: $$value => {
					error = $$value;
					$$settled = false;
				}
			},
			{}
		)}
    ${validate_component(Card, "Card").$$render(
			$$result,
			{
				data: measurements[2],
				period,
				anyError: error
			},
			{
				anyError: $$value => {
					error = $$value;
					$$settled = false;
				}
			},
			{}
		)}</div></div>
${validate_component(Footer, "Footer").$$render($$result, {}, {}, {})}`;
	} while (!$$settled);

	return $$rendered;
});

/* src\App.svelte generated by Svelte v3.53.1 */

const App = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let { url = "" } = $$props;
	if ($$props.url === void 0 && $$bindings.url && url !== void 0) $$bindings.url(url);

	return `${validate_component(Router, "Router").$$render($$result, { url }, {}, {
		default: () => {
			return `${validate_component(Route, "Route").$$render($$result, { path: "/", component: Home }, {}, {})}
    ${validate_component(Route, "Route").$$render($$result, { path: "/generate", component: Generate }, {}, {})}
    ${validate_component(Route, "Route").$$render($$result, { path: "/dashboard", component: SignIn }, {}, {})}
    ${validate_component(Route, "Route").$$render(
				$$result,
				{
					path: "/dashboard/demo",
					component: Dashboard,
					demo: true,
					userID: null
				},
				{},
				{}
			)}
    ${validate_component(Route, "Route").$$render(
				$$result,
				{
					path: "/dashboard/:userID",
					component: Dashboard,
					demo: false
				},
				{},
				{}
			)}
    ${validate_component(Route, "Route").$$render(
				$$result,
				{
					path: "/monitoring/:userID",
					component: Monitoring
				},
				{},
				{}
			)}
    ${validate_component(Route, "Route").$$render($$result, { path: "/delete", component: Delete }, {}, {})}`;
		}
	})}`;
});

module.exports = App;
