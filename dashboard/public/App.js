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

/* src\routes\Generate.svelte generated by Svelte v3.53.1 */

const css$a = {
	code: ".generate.svelte-x7dl2x{display:grid;place-items:center}h2.svelte-x7dl2x{font-family:-apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, Oxygen,\n      Ubuntu, Cantarell, \"Open Sans\", \"Helvetica Neue\", sans-serif;margin:0 0 1em;font-size:2em}.content.svelte-x7dl2x{width:fit-content;background:#343434;background:#1c1c1c;padding:3.5em 4em 4em;border-radius:9px;margin:20vh 0 2vh}input.svelte-x7dl2x{background:#1c1c1c;background:#343434;border:none;padding:0 20px;width:300px;font-size:1em;text-align:center;height:40px;border-radius:4px;margin-bottom:2.5em;color:white;display:grid}button.svelte-x7dl2x{height:40px;border-radius:4px;padding:0 20px;border:none;cursor:pointer;width:100px}.highlight.svelte-x7dl2x{color:#3fcf8e}.details.svelte-x7dl2x{font-size:0.8em}.keep-secure.svelte-x7dl2x{color:#5a5a5a;margin-bottom:1em}#copyBtn.svelte-x7dl2x{background:#1c1c1c;display:none;background:#343434;place-items:center;margin:auto}.copy-icon.svelte-x7dl2x{filter:contrast(0.3);height:20px}#copied.svelte-x7dl2x{color:var(--highlight);margin:2em auto auto;visibility:hidden;height:1em}.spinner.svelte-x7dl2x{height:7em;margin-bottom:5em}#generateBtn.svelte-x7dl2x{background:#3fcf8e}",
	map: "{\"version\":3,\"file\":\"Generate.svelte\",\"sources\":[\"Generate.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">let loading = false;\\r\\nlet generatedKey = false;\\r\\nlet apiKey = \\\"\\\";\\r\\nlet generateBtn;\\r\\nlet copyBtn;\\r\\nlet copiedNotification;\\r\\nasync function genAPIKey() {\\r\\n    if (!generatedKey) {\\r\\n        loading = true;\\r\\n        const response = await fetch(\\\"https://api-analytics-server.vercel.app/api/generate-api-key\\\");\\r\\n        if (response.status == 200) {\\r\\n            const data = await response.json();\\r\\n            generatedKey = true;\\r\\n            apiKey = data.value;\\r\\n            generateBtn.style.display = \\\"none\\\";\\r\\n            copyBtn.style.display = \\\"grid\\\";\\r\\n        }\\r\\n        loading = false;\\r\\n    }\\r\\n}\\r\\nfunction copyToClipboard() {\\r\\n    navigator.clipboard.writeText(apiKey);\\r\\n    copyBtn.value = \\\"Copied\\\";\\r\\n    copiedNotification.style.visibility = \\\"visible\\\";\\r\\n}\\r\\n</script>\\n\\n<div class=\\\"generate\\\">\\n  <div class=\\\"content\\\">\\n    <h2>Generate API key</h2>\\n    <input type=\\\"text\\\" readonly bind:value={apiKey} />\\n    <button id=\\\"generateBtn\\\" on:click={genAPIKey} bind:this={generateBtn}\\n      >Generate</button\\n    >\\n    <button id=\\\"copyBtn\\\" on:click={copyToClipboard} bind:this={copyBtn}\\n      ><img class=\\\"copy-icon\\\" src=\\\"img/copy.png\\\" alt=\\\"\\\" /></button\\n    >\\n    <div id=\\\"copied\\\" bind:this={copiedNotification}>Copied!</div>\\n\\n    <div class=\\\"spinner\\\">\\n      <div class=\\\"loader\\\" style=\\\"display: {loading ? 'initial' : 'none'}\\\" />\\n    </div>\\n\\n    <div class=\\\"details\\\">\\n      <div class=\\\"keep-secure\\\">Keep your API key safe and secure.</div>\\n      <div class=\\\"highlight logo\\\">API Analytics</div>\\n    </div>\\n  </div>\\n</div>\\n\\n<style>\\n  .generate {\\n    display: grid;\\n    place-items: center;\\n  }\\n  h2 {\\n    font-family: -apple-system, BlinkMacSystemFont, \\\"Segoe UI\\\", Roboto, Oxygen,\\n      Ubuntu, Cantarell, \\\"Open Sans\\\", \\\"Helvetica Neue\\\", sans-serif;\\n    margin: 0 0 1em;\\n    font-size: 2em;\\n  }\\n  .content {\\n    width: fit-content;\\n    background: #343434;\\n    background: #1c1c1c;\\n    padding: 3.5em 4em 4em;\\n    border-radius: 9px;\\n    margin: 20vh 0 2vh;\\n  }\\n  input {\\n    background: #1c1c1c;\\n    background: #343434;\\n    border: none;\\n    padding: 0 20px;\\n    width: 300px;\\n    font-size: 1em;\\n    text-align: center;\\n    height: 40px;\\n    border-radius: 4px;\\n    margin-bottom: 2.5em;\\n    color: white;\\n    display: grid;\\n  }\\n  button {\\n    height: 40px;\\n    border-radius: 4px;\\n    padding: 0 20px;\\n    border: none;\\n    cursor: pointer;\\n    width: 100px;\\n  }\\n  .highlight {\\n    color: #3fcf8e;\\n  }\\n  .details {\\n    font-size: 0.8em;\\n  }\\n  .keep-secure {\\n    color: #5a5a5a;\\n    margin-bottom: 1em;\\n  }\\n\\n  #copyBtn {\\n    background: #1c1c1c;\\n    display: none;\\n    background: #343434;\\n    place-items: center;\\n    margin: auto;\\n  }\\n  .copy-icon {\\n    filter: contrast(0.3);\\n    height: 20px;\\n  }\\n  #copied {\\n    color: var(--highlight);\\n    margin: 2em auto auto;\\n    visibility: hidden;\\n    height: 1em;\\n  }\\n\\n  .spinner {\\n    height: 7em;\\n    margin-bottom: 5em;\\n  }\\n  #generateBtn {\\n    background: #3fcf8e;\\n  }\\n\\n</style>\\n\"],\"names\":[],\"mappings\":\"AAmDE,SAAS,cAAC,CAAC,AACT,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,EAAE,cAAC,CAAC,AACF,WAAW,CAAE,aAAa,CAAC,CAAC,kBAAkB,CAAC,CAAC,UAAU,CAAC,CAAC,MAAM,CAAC,CAAC,MAAM,CAAC;MACzE,MAAM,CAAC,CAAC,SAAS,CAAC,CAAC,WAAW,CAAC,CAAC,gBAAgB,CAAC,CAAC,UAAU,CAC9D,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CACf,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,KAAK,CAAE,WAAW,CAClB,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,KAAK,CAAC,GAAG,CAAC,GAAG,CACtB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,CAAC,CAAC,CAAC,GAAG,AACpB,CAAC,AACD,KAAK,cAAC,CAAC,AACL,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,MAAM,CAAE,IAAI,CACZ,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,KAAK,CAAE,KAAK,CACZ,SAAS,CAAE,GAAG,CACd,UAAU,CAAE,MAAM,CAClB,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,aAAa,CAAE,KAAK,CACpB,KAAK,CAAE,KAAK,CACZ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,cAAC,CAAC,AACN,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,OAAO,CACf,KAAK,CAAE,KAAK,AACd,CAAC,AACD,UAAU,cAAC,CAAC,AACV,KAAK,CAAE,OAAO,AAChB,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,YAAY,cAAC,CAAC,AACZ,KAAK,CAAE,OAAO,CACd,aAAa,CAAE,GAAG,AACpB,CAAC,AAED,QAAQ,cAAC,CAAC,AACR,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,IAAI,CACb,UAAU,CAAE,OAAO,CACnB,WAAW,CAAE,MAAM,CACnB,MAAM,CAAE,IAAI,AACd,CAAC,AACD,UAAU,cAAC,CAAC,AACV,MAAM,CAAE,SAAS,GAAG,CAAC,CACrB,MAAM,CAAE,IAAI,AACd,CAAC,AACD,OAAO,cAAC,CAAC,AACP,KAAK,CAAE,IAAI,WAAW,CAAC,CACvB,MAAM,CAAE,GAAG,CAAC,IAAI,CAAC,IAAI,CACrB,UAAU,CAAE,MAAM,CAClB,MAAM,CAAE,GAAG,AACb,CAAC,AAED,QAAQ,cAAC,CAAC,AACR,MAAM,CAAE,GAAG,CACX,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,YAAY,cAAC,CAAC,AACZ,UAAU,CAAE,OAAO,AACrB,CAAC\"}"
};

const Generate = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let apiKey = "";
	let generateBtn;
	let copyBtn;
	let copiedNotification;

	$$result.css.add(css$a);

	return `<div class="${"generate svelte-x7dl2x"}"><div class="${"content svelte-x7dl2x"}"><h2 class="${"svelte-x7dl2x"}">Generate API key</h2>
    <input type="${"text"}" readonly class="${"svelte-x7dl2x"}"${add_attribute("value", apiKey, 0)}>
    <button id="${"generateBtn"}" class="${"svelte-x7dl2x"}"${add_attribute("this", generateBtn, 0)}>Generate</button>
    <button id="${"copyBtn"}" class="${"svelte-x7dl2x"}"${add_attribute("this", copyBtn, 0)}><img class="${"copy-icon svelte-x7dl2x"}" src="${"img/copy.png"}" alt="${""}"></button>
    <div id="${"copied"}" class="${"svelte-x7dl2x"}"${add_attribute("this", copiedNotification, 0)}>Copied!</div>

    <div class="${"spinner svelte-x7dl2x"}"><div class="${"loader"}" style="${"display: " + escape('none', true)}"></div></div>

    <div class="${"details svelte-x7dl2x"}"><div class="${"keep-secure svelte-x7dl2x"}">Keep your API key safe and secure.</div>
      <div class="${"highlight logo svelte-x7dl2x"}">API Analytics</div></div></div>
</div>`;
});

/* src\routes\SignIn.svelte generated by Svelte v3.53.1 */

const css$9 = {
	code: ".generate.svelte-1nnnhku{display:grid;place-items:center}h2.svelte-1nnnhku{font-family:-apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, Oxygen,\n      Ubuntu, Cantarell, \"Open Sans\", \"Helvetica Neue\", sans-serif;margin:0 0 1em;font-size:2em}.content.svelte-1nnnhku{width:fit-content;background:#343434;background:#1c1c1c;padding:3.5em 4em 4em;border-radius:9px;margin:20vh 0 2vh}input.svelte-1nnnhku{background:#1c1c1c;background:#343434;border:none;padding:0 20px;width:300px;font-size:1em;text-align:center;height:40px;border-radius:4px;margin-bottom:2.5em;color:white;display:grid}button.svelte-1nnnhku{height:40px;border-radius:4px;padding:0 20px;border:none;cursor:pointer;width:100px}.highlight.svelte-1nnnhku{color:#3fcf8e}.details.svelte-1nnnhku{font-size:0.8em}.keep-secure.svelte-1nnnhku{color:#5a5a5a;margin-bottom:1em}#generateBtn.svelte-1nnnhku{background:#3fcf8e}",
	map: "{\"version\":3,\"file\":\"SignIn.svelte\",\"sources\":[\"SignIn.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">let apiKey = \\\"\\\";\\r\\nlet loading = false;\\r\\nasync function genAPIKey() {\\r\\n    loading = true;\\r\\n    // Fetch page ID\\r\\n    const response = await fetch(`https://api-analytics-server.vercel.app/api/user-id/${apiKey}`);\\r\\n    console.log(response);\\r\\n    if (response.status == 200) {\\r\\n        const data = await response.json();\\r\\n        window.location.href = `/dashboard/${data.value.replaceAll(\\\"-\\\", \\\"\\\")}`;\\r\\n    }\\r\\n    loading = false;\\r\\n}\\r\\n</script>\\n\\n<div class=\\\"generate\\\">\\n  <div class=\\\"content\\\">\\n    <h2>Dashboard</h2>\\n    <input type=\\\"text\\\" bind:value={apiKey} placeholder=\\\"Enter API key\\\"/>\\n    <button id=\\\"generateBtn\\\" on:click={genAPIKey}>Load</button>\\n    <div class=\\\"spinner\\\">\\n      <div class=\\\"loader\\\" style=\\\"display: {loading ? 'initial' : 'none'}\\\" />\\n    </div>\\n    <div class=\\\"details\\\">\\n      <div class=\\\"keep-secure\\\">Keep your API key safe and secure.</div>\\n      <div class=\\\"highlight logo\\\">API Analytics</div>\\n    </div>\\n  </div>\\n</div>\\n\\n<style>\\n  .generate {\\n    display: grid;\\n    place-items: center;\\n  }\\n  h2 {\\n    font-family: -apple-system, BlinkMacSystemFont, \\\"Segoe UI\\\", Roboto, Oxygen,\\n      Ubuntu, Cantarell, \\\"Open Sans\\\", \\\"Helvetica Neue\\\", sans-serif;\\n    margin: 0 0 1em;\\n    font-size: 2em;\\n  }\\n  .content {\\n    width: fit-content;\\n    background: #343434;\\n    background: #1c1c1c;\\n    padding: 3.5em 4em 4em;\\n    border-radius: 9px;\\n    margin: 20vh 0 2vh;\\n  }\\n  input {\\n    background: #1c1c1c;\\n    background: #343434;\\n    border: none;\\n    padding: 0 20px;\\n    width: 300px;\\n    font-size: 1em;\\n    text-align: center;\\n    height: 40px;\\n    border-radius: 4px;\\n    margin-bottom: 2.5em;\\n    color: white;\\n    display: grid;\\n  }\\n  button {\\n    height: 40px;\\n    border-radius: 4px;\\n    padding: 0 20px;\\n    border: none;\\n    cursor: pointer;\\n    width: 100px;\\n  }\\n  .highlight {\\n    color: #3fcf8e;\\n  }\\n  .details {\\n    /* margin-top: 15rem; */\\n    font-size: 0.8em;\\n  }\\n  .keep-secure {\\n    color: #5a5a5a;\\n    margin-bottom: 1em;\\n  }\\n  #generateBtn {\\n    background: #3fcf8e;\\n  }\\n\\n</style>\\n\"],\"names\":[],\"mappings\":\"AA+BE,SAAS,eAAC,CAAC,AACT,OAAO,CAAE,IAAI,CACb,WAAW,CAAE,MAAM,AACrB,CAAC,AACD,EAAE,eAAC,CAAC,AACF,WAAW,CAAE,aAAa,CAAC,CAAC,kBAAkB,CAAC,CAAC,UAAU,CAAC,CAAC,MAAM,CAAC,CAAC,MAAM,CAAC;MACzE,MAAM,CAAC,CAAC,SAAS,CAAC,CAAC,WAAW,CAAC,CAAC,gBAAgB,CAAC,CAAC,UAAU,CAC9D,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,GAAG,CACf,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,KAAK,CAAE,WAAW,CAClB,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,OAAO,CAAE,KAAK,CAAC,GAAG,CAAC,GAAG,CACtB,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,IAAI,CAAC,CAAC,CAAC,GAAG,AACpB,CAAC,AACD,KAAK,eAAC,CAAC,AACL,UAAU,CAAE,OAAO,CACnB,UAAU,CAAE,OAAO,CACnB,MAAM,CAAE,IAAI,CACZ,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,KAAK,CAAE,KAAK,CACZ,SAAS,CAAE,GAAG,CACd,UAAU,CAAE,MAAM,CAClB,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,aAAa,CAAE,KAAK,CACpB,KAAK,CAAE,KAAK,CACZ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,eAAC,CAAC,AACN,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAClB,OAAO,CAAE,CAAC,CAAC,IAAI,CACf,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,OAAO,CACf,KAAK,CAAE,KAAK,AACd,CAAC,AACD,UAAU,eAAC,CAAC,AACV,KAAK,CAAE,OAAO,AAChB,CAAC,AACD,QAAQ,eAAC,CAAC,AAER,SAAS,CAAE,KAAK,AAClB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,KAAK,CAAE,OAAO,CACd,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,YAAY,eAAC,CAAC,AACZ,UAAU,CAAE,OAAO,AACrB,CAAC\"}"
};

const SignIn = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let apiKey = "";

	$$result.css.add(css$9);

	return `<div class="${"generate svelte-1nnnhku"}"><div class="${"content svelte-1nnnhku"}"><h2 class="${"svelte-1nnnhku"}">Dashboard</h2>
    <input type="${"text"}" placeholder="${"Enter API key"}" class="${"svelte-1nnnhku"}"${add_attribute("value", apiKey, 0)}>
    <button id="${"generateBtn"}" class="${"svelte-1nnnhku"}">Load</button>
    <div class="${"spinner"}"><div class="${"loader"}" style="${"display: " + escape('none', true)}"></div></div>
    <div class="${"details svelte-1nnnhku"}"><div class="${"keep-secure svelte-1nnnhku"}">Keep your API key safe and secure.</div>
      <div class="${"highlight logo svelte-1nnnhku"}">API Analytics</div></div></div>
</div>`;
});

/* src\components\Requests.svelte generated by Svelte v3.53.1 */

const css$8 = {
	code: ".card.svelte-1bbcp2r{width:calc(200px - 1em);margin:0 1em 0 2em}.value.svelte-1bbcp2r{margin:20px 0;font-size:1.8em;font-weight:600}.per-hour.svelte-1bbcp2r{color:var(--dim-text);font-size:0.8em;margin-left:4px}",
	map: "{\"version\":3,\"file\":\"Requests.svelte\",\"sources\":[\"Requests.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction thisWeek(date) {\\r\\n    let weekAgo = new Date();\\r\\n    weekAgo.setDate(weekAgo.getDate() - 7);\\r\\n    return date > weekAgo;\\r\\n}\\r\\nfunction build() {\\r\\n    let totalRequests = 0;\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        if (thisWeek(date)) {\\r\\n            totalRequests++;\\r\\n        }\\r\\n    }\\r\\n    requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);\\r\\n}\\r\\nlet requestsPerHour;\\r\\nonMount(() => {\\r\\n    build();\\r\\n});\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Last week\\\">\\r\\n  <div class=\\\"card-title\\\">\\r\\n    Requests <span class=\\\"per-hour\\\">/ hour</span>\\r\\n  </div>\\r\\n  {#if requestsPerHour != undefined}\\r\\n    <div class=\\\"value\\\">{requestsPerHour}</div>\\r\\n  {/if}\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    width: calc(200px - 1em);\\r\\n    margin: 0 1em 0 2em;\\r\\n  }\\r\\n  .value {\\r\\n    margin: 20px 0;\\r\\n    font-size: 1.8em;\\r\\n    font-weight: 600;\\r\\n  }\\r\\n  .per-hour {\\r\\n    color: var(--dim-text);\\r\\n    font-size: 0.8em;\\r\\n    margin-left: 4px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAiCE,KAAK,eAAC,CAAC,AACL,KAAK,CAAE,KAAK,KAAK,CAAC,CAAC,CAAC,GAAG,CAAC,CACxB,MAAM,CAAE,CAAC,CAAC,GAAG,CAAC,CAAC,CAAC,GAAG,AACrB,CAAC,AACD,MAAM,eAAC,CAAC,AACN,MAAM,CAAE,IAAI,CAAC,CAAC,CACd,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AACD,SAAS,eAAC,CAAC,AACT,KAAK,CAAE,IAAI,UAAU,CAAC,CACtB,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC\"}"
};

function thisWeek(date) {
	let weekAgo = new Date();
	weekAgo.setDate(weekAgo.getDate() - 7);
	return date > weekAgo;
}

const Requests = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function build() {
		let totalRequests = 0;

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);

			if (thisWeek(date)) {
				totalRequests++;
			}
		}

		requestsPerHour = (24 * 7 / totalRequests).toFixed(2);
	}

	let requestsPerHour;

	onMount(() => {
		build();
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$8);

	return `<div class="${"card svelte-1bbcp2r"}" title="${"Last week"}"><div class="${"card-title"}">Requests <span class="${"per-hour svelte-1bbcp2r"}">/ hour</span></div>
  ${requestsPerHour != undefined
	? `<div class="${"value svelte-1bbcp2r"}">${escape(requestsPerHour)}</div>`
	: ``}
</div>`;
});

/* src\components\ResponseTimes.svelte generated by Svelte v3.53.1 */

const css$7 = {
	code: ".values.svelte-kx5j01{display:flex;color:var(--highlight);font-size:1.8em;font-weight:700}.values.svelte-kx5j01,.labels.svelte-kx5j01{margin:0 0.5rem}.value.svelte-kx5j01{flex:1;font-size:1.1em;padding:20px 20px 4px}.labels.svelte-kx5j01{display:flex;font-size:0.8em;color:var(--dim-text)}.label.svelte-kx5j01{flex:1}.milliseconds.svelte-kx5j01{color:var(--dim-text);font-size:0.8em;margin-left:4px}.median.svelte-kx5j01{font-size:1em}.upper-quartile.svelte-kx5j01,.lower-quartile.svelte-kx5j01{font-size:1em;padding-bottom:0}.bar.svelte-kx5j01{padding:20px 0 20px;display:flex;height:30px;width:85%;margin:auto;align-items:center;position:relative}.bar-green.svelte-kx5j01{background:var(--highlight);width:75%;height:10px;border-radius:2px 0 0 2px}.bar-yellow.svelte-kx5j01{width:15%;height:10px;background:rgb(235, 235, 129)}.bar-red.svelte-kx5j01{width:20%;height:10px;border-radius:0 2px 2px 0;background:rgb(228, 97, 97)}.marker.svelte-kx5j01{position:absolute;height:30px;width:5px;background:white;border-radius:2px;left:0}",
	map: "{\"version\":3,\"file\":\"ResponseTimes.svelte\",\"sources\":[\"ResponseTimes.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\n// Median and quartiles from StackOverflow answer\\r\\n// https://stackoverflow.com/a/55297611/8851732\\r\\nconst asc = (arr) => arr.sort((a, b) => a - b);\\r\\nconst sum = (arr) => arr.reduce((a, b) => a + b, 0);\\r\\nconst mean = (arr) => sum(arr) / arr.length;\\r\\n// sample standard deviation\\r\\nfunction std(arr) {\\r\\n    const mu = mean(arr);\\r\\n    const diffArr = arr.map((a) => (a - mu) ** 2);\\r\\n    return Math.sqrt(sum(diffArr) / (arr.length - 1));\\r\\n}\\r\\nfunction quantile(arr, q) {\\r\\n    const sorted = asc(arr);\\r\\n    const pos = (sorted.length - 1) * q;\\r\\n    const base = Math.floor(pos);\\r\\n    const rest = pos - base;\\r\\n    if (sorted[base + 1] !== undefined) {\\r\\n        return sorted[base] + rest * (sorted[base + 1] - sorted[base]);\\r\\n    }\\r\\n    else {\\r\\n        return sorted[base];\\r\\n    }\\r\\n}\\r\\nfunction markerPosition(x) {\\r\\n    let position = Math.log10(x) * 125 - 300;\\r\\n    if (position < 0) {\\r\\n        return 0;\\r\\n    }\\r\\n    else if (position > 100) {\\r\\n        return 100;\\r\\n    }\\r\\n    else {\\r\\n        return position;\\r\\n    }\\r\\n}\\r\\nfunction setMarkerPosition(median) {\\r\\n    let position = markerPosition(median);\\r\\n    marker.style.left = `${position}%`;\\r\\n}\\r\\nfunction build() {\\r\\n    let responseTimes = [];\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        responseTimes.push(data[i].response_time);\\r\\n    }\\r\\n    median = quantile(responseTimes, 0.25);\\r\\n    LQ = quantile(responseTimes, 0.5);\\r\\n    UQ = quantile(responseTimes, 0.75);\\r\\n    setMarkerPosition(median);\\r\\n}\\r\\nlet median, LQ, UQ;\\r\\nlet marker;\\r\\nonMount(() => {\\r\\n    build();\\r\\n});\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">\\r\\n    Response Times <span class=\\\"milliseconds\\\">(ms)</span>\\r\\n  </div>\\r\\n  <div class=\\\"values\\\">\\r\\n    <div class=\\\"value lower-quartile\\\">{LQ}</div>\\r\\n    <div class=\\\"value median\\\">{median}</div>\\r\\n    <div class=\\\"value upper-quartile\\\">{UQ}</div>\\r\\n  </div>\\r\\n  <div class=\\\"labels\\\">\\r\\n    <div class=\\\"label\\\">25%</div>\\r\\n    <div class=\\\"label\\\">Median</div>\\r\\n    <div class=\\\"label\\\">75%</div>\\r\\n  </div>\\r\\n  <div class=\\\"bar\\\">\\r\\n    <div class=\\\"bar-green\\\" />\\r\\n    <div class=\\\"bar-yellow\\\" />\\r\\n    <div class=\\\"bar-red\\\" />\\r\\n    <div class=\\\"marker\\\" bind:this={marker} />\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .values {\\r\\n    display: flex;\\r\\n    color: var(--highlight);\\r\\n    font-size: 1.8em;\\r\\n    font-weight: 700;\\r\\n  }\\r\\n  .values,\\r\\n  .labels {\\r\\n    margin: 0 0.5rem;\\r\\n  }\\r\\n  .value {\\r\\n    flex: 1;\\r\\n    font-size: 1.1em;\\r\\n    padding: 20px 20px 4px;\\r\\n  }\\r\\n  .labels {\\r\\n    display: flex;\\r\\n    font-size: 0.8em;\\r\\n    color: var(--dim-text);\\r\\n  }\\r\\n  .label {\\r\\n    flex: 1;\\r\\n  }\\r\\n\\r\\n  .milliseconds {\\r\\n    color: var(--dim-text);\\r\\n    font-size: 0.8em;\\r\\n    margin-left: 4px;\\r\\n  }\\r\\n\\r\\n  .median {\\r\\n    font-size: 1em;\\r\\n  }\\r\\n  .upper-quartile,\\r\\n  .lower-quartile {\\r\\n    font-size: 1em;\\r\\n    padding-bottom: 0;\\r\\n  }\\r\\n\\r\\n  .bar {\\r\\n    padding: 20px 0 20px;\\r\\n    display: flex;\\r\\n    height: 30px;\\r\\n    width: 85%;\\r\\n    margin: auto;\\r\\n    align-items: center;\\r\\n    position: relative;\\r\\n  }\\r\\n  .bar-green {\\r\\n    background: var(--highlight);\\r\\n    width: 75%;\\r\\n    height: 10px;\\r\\n    border-radius: 2px 0 0 2px;\\r\\n  }\\r\\n  .bar-yellow {\\r\\n    width: 15%;\\r\\n    height: 10px;\\r\\n    background: rgb(235, 235, 129);\\r\\n  }\\r\\n  .bar-red {\\r\\n    width: 20%;\\r\\n    height: 10px;\\r\\n    border-radius: 0 2px 2px 0;\\r\\n    background: rgb(228, 97, 97);\\r\\n  }\\r\\n  .marker {\\r\\n    position: absolute;\\r\\n    height: 30px;\\r\\n    width: 5px;\\r\\n    background: white;\\r\\n    border-radius: 2px;\\r\\n    left: 0; /* Changed during runtime to reflect median */\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAiFE,OAAO,cAAC,CAAC,AACP,OAAO,CAAE,IAAI,CACb,KAAK,CAAE,IAAI,WAAW,CAAC,CACvB,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AACD,qBAAO,CACP,OAAO,cAAC,CAAC,AACP,MAAM,CAAE,CAAC,CAAC,MAAM,AAClB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,IAAI,CAAE,CAAC,CACP,SAAS,CAAE,KAAK,CAChB,OAAO,CAAE,IAAI,CAAC,IAAI,CAAC,GAAG,AACxB,CAAC,AACD,OAAO,cAAC,CAAC,AACP,OAAO,CAAE,IAAI,CACb,SAAS,CAAE,KAAK,CAChB,KAAK,CAAE,IAAI,UAAU,CAAC,AACxB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,IAAI,CAAE,CAAC,AACT,CAAC,AAED,aAAa,cAAC,CAAC,AACb,KAAK,CAAE,IAAI,UAAU,CAAC,CACtB,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,AAClB,CAAC,AAED,OAAO,cAAC,CAAC,AACP,SAAS,CAAE,GAAG,AAChB,CAAC,AACD,6BAAe,CACf,eAAe,cAAC,CAAC,AACf,SAAS,CAAE,GAAG,CACd,cAAc,CAAE,CAAC,AACnB,CAAC,AAED,IAAI,cAAC,CAAC,AACJ,OAAO,CAAE,IAAI,CAAC,CAAC,CAAC,IAAI,CACpB,OAAO,CAAE,IAAI,CACb,MAAM,CAAE,IAAI,CACZ,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,WAAW,CAAE,MAAM,CACnB,QAAQ,CAAE,QAAQ,AACpB,CAAC,AACD,UAAU,cAAC,CAAC,AACV,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,GAAG,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,AAC5B,CAAC,AACD,WAAW,cAAC,CAAC,AACX,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,UAAU,CAAE,IAAI,GAAG,CAAC,CAAC,GAAG,CAAC,CAAC,GAAG,CAAC,AAChC,CAAC,AACD,QAAQ,cAAC,CAAC,AACR,KAAK,CAAE,GAAG,CACV,MAAM,CAAE,IAAI,CACZ,aAAa,CAAE,CAAC,CAAC,GAAG,CAAC,GAAG,CAAC,CAAC,CAC1B,UAAU,CAAE,IAAI,GAAG,CAAC,CAAC,EAAE,CAAC,CAAC,EAAE,CAAC,AAC9B,CAAC,AACD,OAAO,cAAC,CAAC,AACP,QAAQ,CAAE,QAAQ,CAClB,MAAM,CAAE,IAAI,CACZ,KAAK,CAAE,GAAG,CACV,UAAU,CAAE,KAAK,CACjB,aAAa,CAAE,GAAG,CAClB,IAAI,CAAE,CAAC,AACT,CAAC\"}"
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

		if (sorted[base + 1] !== undefined) {
			return sorted[base] + rest * (sorted[base + 1] - sorted[base]);
		} else {
			return sorted[base];
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

		median = quantile(responseTimes, 0.25);
		LQ = quantile(responseTimes, 0.5);
		UQ = quantile(responseTimes, 0.75);
		setMarkerPosition(median);
	}

	let median, LQ, UQ;
	let marker;

	onMount(() => {
		build();
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$7);

	return `<div class="${"card"}"><div class="${"card-title"}">Response Times <span class="${"milliseconds svelte-kx5j01"}">(ms)</span></div>
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

/* src\components\Endpoints.svelte generated by Svelte v3.53.1 */

const css$6 = {
	code: ".endpoints.svelte-1e031qh{margin:1em 20px 0.6em}.endpoint.svelte-1e031qh{border-radius:3px;margin:6px 0;color:var(--light-background);text-align:left;position:relative}.endpoint-label.svelte-1e031qh{display:flex}.path.svelte-1e031qh,.count.svelte-1e031qh{padding:3px 15px}.path.svelte-1e031qh{flex-grow:1}.endpoint-container.svelte-1e031qh{display:flex}.external-label.svelte-1e031qh{padding:3px 15px;left:40px;top:0;margin:6px 0;color:#707070;display:none}",
	map: "{\"version\":3,\"file\":\"Endpoints.svelte\",\"sources\":[\"Endpoints.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction endpointFreq() {\\r\\n    let freq = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let endpointID = data[i].path + data[i].status;\\r\\n        if (!(endpointID in freq)) {\\r\\n            freq[endpointID] = {\\r\\n                path: data[i].path,\\r\\n                status: data[i].status,\\r\\n                count: 0,\\r\\n            };\\r\\n        }\\r\\n        freq[endpointID].count++;\\r\\n    }\\r\\n    return freq;\\r\\n}\\r\\nfunction build() {\\r\\n    let freq = endpointFreq();\\r\\n    let freqArr = [];\\r\\n    maxCount = 0;\\r\\n    for (let endpointID in freq) {\\r\\n        freqArr.push(freq[endpointID]);\\r\\n        if (freq[endpointID].count > maxCount) {\\r\\n            maxCount = freq[endpointID].count;\\r\\n        }\\r\\n    }\\r\\n    freqArr.sort((a, b) => {\\r\\n        return b.count - a.count;\\r\\n    });\\r\\n    endpoints = freqArr;\\r\\n}\\r\\nfunction setEndpointLabelVisibility(idx) {\\r\\n    let endpoint = document.getElementById(`endpoint-label-${idx}`);\\r\\n    let endpointPath = document.getElementById(`endpoint-path-${idx}`);\\r\\n    let endpointCount = document.getElementById(`endpoint-count-${idx}`);\\r\\n    let externalLabel = document.getElementById(`external-label-${idx}`);\\r\\n    if (endpoint.clientWidth <\\r\\n        endpointPath.clientWidth + endpointCount.clientWidth) {\\r\\n        externalLabel.style.display = \\\"flex\\\";\\r\\n        endpointPath.style.display = \\\"none\\\";\\r\\n    }\\r\\n}\\r\\nfunction setEndpointLabels() {\\r\\n    for (let i = 0; i < endpoints.length; i++) {\\r\\n        setEndpointLabelVisibility(i);\\r\\n    }\\r\\n}\\r\\nonMount(() => {\\r\\n    build();\\r\\n    setTimeout(setEndpointLabels, 50);\\r\\n});\\r\\nlet endpoints;\\r\\nlet maxCount;\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">Endpoints</div>\\r\\n  {#if endpoints != undefined}\\r\\n    <div class=\\\"endpoints\\\">\\r\\n      {#each endpoints as endpoint, i}\\r\\n        <div class=\\\"endpoint-container\\\">\\r\\n          <div\\r\\n            class=\\\"endpoint\\\"\\r\\n            id=\\\"endpoint-{i}\\\"\\r\\n            style=\\\"width: {(endpoint.count / maxCount) *\\r\\n              100}%; background: {endpoint.status >= 200 &&\\r\\n            endpoint.status <= 299\\r\\n              ? 'var(--highlight)'\\r\\n              : '#e46161'}\\\"\\r\\n          >\\r\\n            <div class=\\\"endpoint-label\\\" id=\\\"endpoint-label-{i}\\\">\\r\\n              <div class=\\\"path\\\" id=\\\"endpoint-path-{i}\\\">\\r\\n                {endpoint.path}\\r\\n              </div>\\r\\n              <div class=\\\"count\\\" id=\\\"endpoint-count-{i}\\\">{endpoint.count}</div>\\r\\n            </div>\\r\\n          </div>\\r\\n          <div class=\\\"external-label\\\" id=\\\"external-label-{i}\\\">\\r\\n            <div class=\\\"external-label-path\\\">\\r\\n              {endpoint.path}\\r\\n            </div>\\r\\n          </div>\\r\\n        </div>\\r\\n      {/each}\\r\\n    </div>\\r\\n  {/if}\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .endpoints {\\r\\n    margin: 1em 20px 0.6em;\\r\\n  }\\r\\n  .endpoint {\\r\\n    border-radius: 3px;\\r\\n    margin: 6px 0;\\r\\n    color: var(--light-background);\\r\\n    text-align: left;\\r\\n    position: relative;\\r\\n  }\\r\\n  .endpoint-label {\\r\\n    display: flex;\\r\\n  }\\r\\n  .path,\\r\\n  .count {\\r\\n    padding: 3px 15px;\\r\\n  }\\r\\n  .path {\\r\\n    flex-grow: 1;\\r\\n  }\\r\\n  .endpoint-container {\\r\\n    display: flex;\\r\\n  }\\r\\n  .external-label {\\r\\n    padding: 3px 15px;\\r\\n    left: 40px;\\r\\n    top: 0;\\r\\n    margin: 6px 0;\\r\\n    color: #707070;\\r\\n    display: none;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AA0FE,UAAU,eAAC,CAAC,AACV,MAAM,CAAE,GAAG,CAAC,IAAI,CAAC,KAAK,AACxB,CAAC,AACD,SAAS,eAAC,CAAC,AACT,aAAa,CAAE,GAAG,CAClB,MAAM,CAAE,GAAG,CAAC,CAAC,CACb,KAAK,CAAE,IAAI,kBAAkB,CAAC,CAC9B,UAAU,CAAE,IAAI,CAChB,QAAQ,CAAE,QAAQ,AACpB,CAAC,AACD,eAAe,eAAC,CAAC,AACf,OAAO,CAAE,IAAI,AACf,CAAC,AACD,oBAAK,CACL,MAAM,eAAC,CAAC,AACN,OAAO,CAAE,GAAG,CAAC,IAAI,AACnB,CAAC,AACD,KAAK,eAAC,CAAC,AACL,SAAS,CAAE,CAAC,AACd,CAAC,AACD,mBAAmB,eAAC,CAAC,AACnB,OAAO,CAAE,IAAI,AACf,CAAC,AACD,eAAe,eAAC,CAAC,AACf,OAAO,CAAE,GAAG,CAAC,IAAI,CACjB,IAAI,CAAE,IAAI,CACV,GAAG,CAAE,CAAC,CACN,MAAM,CAAE,GAAG,CAAC,CAAC,CACb,KAAK,CAAE,OAAO,CACd,OAAO,CAAE,IAAI,AACf,CAAC\"}"
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
}

const Endpoints = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function endpointFreq() {
		let freq = {};

		for (let i = 0; i < data.length; i++) {
			let endpointID = data[i].path + data[i].status;

			if (!(endpointID in freq)) {
				freq[endpointID] = {
					path: data[i].path,
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
	}

	function setEndpointLabels() {
		for (let i = 0; i < endpoints.length; i++) {
			setEndpointLabelVisibility(i);
		}
	}

	onMount(() => {
		build();
		setTimeout(setEndpointLabels, 50);
	});

	let endpoints;
	let maxCount;
	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$6);

	return `<div class="${"card"}"><div class="${"card-title"}">Endpoints</div>
  ${endpoints != undefined
	? `<div class="${"endpoints svelte-1e031qh"}">${each(endpoints, (endpoint, i) => {
			return `<div class="${"endpoint-container svelte-1e031qh"}"><div class="${"endpoint svelte-1e031qh"}" id="${"endpoint-" + escape(i, true)}" style="${"width: " + escape(endpoint.count / maxCount * 100, true) + "%; background: " + escape(
				endpoint.status >= 200 && endpoint.status <= 299
				? 'var(--highlight)'
				: '#e46161',
				true
			)}"><div class="${"endpoint-label svelte-1e031qh"}" id="${"endpoint-label-" + escape(i, true)}"><div class="${"path svelte-1e031qh"}" id="${"endpoint-path-" + escape(i, true)}">${escape(endpoint.path)}</div>
              <div class="${"count svelte-1e031qh"}" id="${"endpoint-count-" + escape(i, true)}">${escape(endpoint.count)}</div>
            </div></div>
          <div class="${"external-label svelte-1e031qh"}" id="${"external-label-" + escape(i, true)}"><div class="${"external-label-path"}">${escape(endpoint.path)}
            </div></div>
        </div>`;
		})}</div>`
	: ``}
</div>`;
});

/* src\components\Footer.svelte generated by Svelte v3.53.1 */

const css$5 = {
	code: ".footer.svelte-zlpuxh{font-size:0.9em;color:var(--highlight)}",
	map: "{\"version\":3,\"file\":\"Footer.svelte\",\"sources\":[\"Footer.svelte\"],\"sourcesContent\":[\"<div class=\\\"footer\\\">API Analytics</div>\\r\\n\\r\\n<style>\\r\\n  .footer {\\r\\n    font-size: 0.9em;\\r\\n    color: var(--highlight);\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAGE,OAAO,cAAC,CAAC,AACP,SAAS,CAAE,KAAK,CAChB,KAAK,CAAE,IAAI,WAAW,CAAC,AACzB,CAAC\"}"
};

const Footer = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	$$result.css.add(css$5);
	return `<div class="${"footer svelte-zlpuxh"}">API Analytics</div>`;
});

/* src\components\SuccessRate.svelte generated by Svelte v3.53.1 */

const css$4 = {
	code: ".card.svelte-yksbz4{width:calc(200px - 1em);margin:0 0 0 1em}.value.svelte-yksbz4{margin:20px 0;font-size:1.8em;font-weight:600;color:var(--yellow)}",
	map: "{\"version\":3,\"file\":\"SuccessRate.svelte\",\"sources\":[\"SuccessRate.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction pastWeek(date) {\\r\\n    let weekAgo = new Date();\\r\\n    weekAgo.setDate(weekAgo.getDate() - 7);\\r\\n    return date > weekAgo;\\r\\n}\\r\\nfunction build() {\\r\\n    let totalRequests = 0;\\r\\n    let successfulRequests = 0;\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        if (pastWeek(date)) {\\r\\n            if (data[i].status >= 200 && data[i].status <= 299) {\\r\\n                successfulRequests++;\\r\\n            }\\r\\n            totalRequests++;\\r\\n        }\\r\\n    }\\r\\n    successRate = (successfulRequests / totalRequests) * 100;\\r\\n}\\r\\nlet successRate;\\r\\nonMount(() => {\\r\\n    build();\\r\\n});\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Last week\\\">\\r\\n  <div class=\\\"card-title\\\">Success Rate</div>\\r\\n  {#if successRate != undefined}\\r\\n    <div\\r\\n      class=\\\"value\\\"\\r\\n      style=\\\"color: {successRate <= 75 ? 'var(--red)' : ''}{successRate > 75 &&\\r\\n      successRate < 90\\r\\n        ? 'var(--yellow)'\\r\\n        : ''}{successRate >= 90 ? 'var(--highlight)' : ''}\\\"\\r\\n    >\\r\\n      {successRate.toFixed(1)}%\\r\\n    </div>\\r\\n  {/if}\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    width: calc(200px - 1em);\\r\\n    margin: 0 0 0 1em;\\r\\n  }\\r\\n\\r\\n  .value {\\r\\n    margin: 20px 0;\\r\\n    font-size: 1.8em;\\r\\n    font-weight: 600;\\r\\n    color: var(--yellow);\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AA2CE,KAAK,cAAC,CAAC,AACL,KAAK,CAAE,KAAK,KAAK,CAAC,CAAC,CAAC,GAAG,CAAC,CACxB,MAAM,CAAE,CAAC,CAAC,CAAC,CAAC,CAAC,CAAC,GAAG,AACnB,CAAC,AAED,MAAM,cAAC,CAAC,AACN,MAAM,CAAE,IAAI,CAAC,CAAC,CACd,SAAS,CAAE,KAAK,CAChB,WAAW,CAAE,GAAG,CAChB,KAAK,CAAE,IAAI,QAAQ,CAAC,AACtB,CAAC\"}"
};

function pastWeek(date) {
	let weekAgo = new Date();
	weekAgo.setDate(weekAgo.getDate() - 7);
	return date > weekAgo;
}

const SuccessRate = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	function build() {
		let totalRequests = 0;
		let successfulRequests = 0;

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);

			if (pastWeek(date)) {
				if (data[i].status >= 200 && data[i].status <= 299) {
					successfulRequests++;
				}

				totalRequests++;
			}
		}

		successRate = successfulRequests / totalRequests * 100;
	}

	let successRate;

	onMount(() => {
		build();
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$4);

	return `<div class="${"card svelte-yksbz4"}" title="${"Last week"}"><div class="${"card-title"}">Success Rate</div>
  ${successRate != undefined
	? `<div class="${"value svelte-yksbz4"}" style="${"color: " + escape(successRate <= 75 ? 'var(--red)' : '', true) + escape(
			successRate > 75 && successRate < 90
			? 'var(--yellow)'
			: '',
			true
		) + escape(successRate >= 90 ? 'var(--highlight)' : '', true)}">${escape(successRate.toFixed(1))}%
    </div>`
	: ``}
</div>`;
});

/* src\components\PastMonth.svelte generated by Svelte v3.53.1 */

const css$3 = {
	code: ".card.svelte-ik169t{margin:0;width:100%}.errors.svelte-ik169t{display:flex;margin-top:8px}.error.svelte-ik169t{background:var(--highlight);flex:1;height:40px;margin:0 1px;border-radius:1px}.success-rate-container.svelte-ik169t{text-align:left;font-size:0.9em;color:#707070}.success-rate-container.svelte-ik169t{margin:1.5em 2em 2em}",
	map: "{\"version\":3,\"file\":\"PastMonth.svelte\",\"sources\":[\"PastMonth.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction pastMonth(date) {\\r\\n    let monthAgo = new Date();\\r\\n    monthAgo.setDate(monthAgo.getDate() - 30);\\r\\n    return date > monthAgo;\\r\\n}\\r\\nlet colors = [\\r\\n    \\\"#444444\\\",\\r\\n    \\\"#E46161\\\",\\r\\n    \\\"#F18359\\\",\\r\\n    \\\"#F5A65A\\\",\\r\\n    \\\"#F3C966\\\",\\r\\n    \\\"#EBEB81\\\",\\r\\n    \\\"#C7E57D\\\",\\r\\n    \\\"#A1DF7E\\\",\\r\\n    \\\"#77D884\\\",\\r\\n    \\\"#3FCF8E\\\", // Green\\r\\n];\\r\\nfunction daysAgo(date) {\\r\\n    let now = new Date();\\r\\n    return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));\\r\\n}\\r\\nfunction setSuccessRate() {\\r\\n    let success = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        if (pastMonth(date)) {\\r\\n            date.setHours(0, 0, 0, 0);\\r\\n            // @ts-ignore\\r\\n            if (!(date in success)) {\\r\\n                // @ts-ignore\\r\\n                success[date] = { total: 0, successful: 0 };\\r\\n            }\\r\\n            if (data[i].status >= 200 && data[i].status <= 299) {\\r\\n                // @ts-ignore\\r\\n                success[date].successful++;\\r\\n            }\\r\\n            // @ts-ignore\\r\\n            success[date].total++;\\r\\n        }\\r\\n    }\\r\\n    let successArr = new Array(60).fill(-0.1); // -0.1 -> 0\\r\\n    for (let date in success) {\\r\\n        let idx = daysAgo(new Date(date));\\r\\n        successArr[successArr.length - idx] = success[date].successful / success[date].total;\\r\\n    }\\r\\n    successRate = successArr;\\r\\n}\\r\\nfunction responseTimePlotLayout() {\\r\\n    let monthAgo = new Date();\\r\\n    monthAgo.setDate(monthAgo.getDate() - 30);\\r\\n    let tomorrow = new Date();\\r\\n    tomorrow.setDate(tomorrow.getDate() + 1);\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 150,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Response time (ms)\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            title: { text: \\\"Date\\\" },\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n            range: [monthAgo, tomorrow],\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nfunction responseTimeLine() {\\r\\n    let points = [];\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        points.push([new Date(data[i].created_at), data[i].response_time]);\\r\\n    }\\r\\n    points.sort((a, b) => {\\r\\n        return a[0] - b[0];\\r\\n    });\\r\\n    let dates = [];\\r\\n    let responses = [];\\r\\n    for (let i = 0; i < points.length; i++) {\\r\\n        dates.push(points[i][0]);\\r\\n        responses.push(points[i][1]);\\r\\n    }\\r\\n    return [\\r\\n        {\\r\\n            x: dates,\\r\\n            y: responses,\\r\\n            mode: \\\"lines\\\",\\r\\n            line: { color: \\\"#3fcf8e\\\" },\\r\\n            hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y}</b><extra></extra>`,\\r\\n            showlegend: false,\\r\\n        },\\r\\n    ];\\r\\n}\\r\\nfunction responseTimePlotData() {\\r\\n    return {\\r\\n        data: responseTimeLine(),\\r\\n        layout: responseTimePlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genResponseTimePlot() {\\r\\n    let plotData = responseTimePlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(responseTimePlotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\nfunction requestsFreqPlotLayout() {\\r\\n    let monthAgo = new Date();\\r\\n    monthAgo.setDate(monthAgo.getDate() - 30);\\r\\n    let tomorrow = new Date();\\r\\n    tomorrow.setDate(tomorrow.getDate() + 1);\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 150,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Requests\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            title: { text: \\\"Date\\\" },\\r\\n            fixedrange: true,\\r\\n            range: [monthAgo, tomorrow],\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nfunction requestsFreqLine() {\\r\\n    let requestFreq = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let date = new Date(data[i].created_at);\\r\\n        date.setHours(0, 0, 0, 0);\\r\\n        // @ts-ignore\\r\\n        if (!(date in requestFreq)) {\\r\\n            // @ts-ignore\\r\\n            requestFreq[date] = 0;\\r\\n        }\\r\\n        // @ts-ignore\\r\\n        requestFreq[date]++;\\r\\n    }\\r\\n    let requestFreqArr = [];\\r\\n    for (let date in requestFreq) {\\r\\n        requestFreqArr.push([new Date(date), requestFreq[date]]);\\r\\n    }\\r\\n    requestFreqArr.sort((a, b) => {\\r\\n        return a[0] - b[0];\\r\\n    });\\r\\n    let dates = [];\\r\\n    let requests = [];\\r\\n    for (let i = 0; i < requestFreqArr.length; i++) {\\r\\n        dates.push(requestFreqArr[i][0]);\\r\\n        requests.push(requestFreqArr[i][1]);\\r\\n    }\\r\\n    return [\\r\\n        {\\r\\n            x: dates,\\r\\n            y: requests,\\r\\n            type: \\\"bar\\\",\\r\\n            marker: { color: \\\"#3fcf8e\\\" },\\r\\n            hovertemplate: `<b>%{y} requests</b><br>%{x|%d %b %Y}</b><extra></extra>`,\\r\\n            showlegend: false,\\r\\n        },\\r\\n    ];\\r\\n}\\r\\nfunction requestsFreqPlotData() {\\r\\n    return {\\r\\n        data: requestsFreqLine(),\\r\\n        layout: requestsFreqPlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genRequestsFreqPlot() {\\r\\n    let plotData = requestsFreqPlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(requestsFreqPlotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\nfunction build() {\\r\\n    setSuccessRate();\\r\\n    genResponseTimePlot();\\r\\n    genRequestsFreqPlot();\\r\\n}\\r\\nlet successRate;\\r\\nlet responseTimePlotDiv;\\r\\nlet requestsFreqPlotDiv;\\r\\nonMount(() => {\\r\\n    build();\\r\\n});\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\">\\r\\n  <div class=\\\"card-title\\\">Past Month</div>\\r\\n  <div id=\\\"plotly\\\">\\r\\n    <div id=\\\"plotDiv\\\" bind:this={requestsFreqPlotDiv}>\\r\\n      <!-- Plotly chart will be drawn inside this DIV -->\\r\\n    </div>\\r\\n  </div>\\r\\n  <div id=\\\"plotly\\\">\\r\\n    <div id=\\\"plotDiv\\\" bind:this={responseTimePlotDiv}>\\r\\n      <!-- Plotly chart will be drawn inside this DIV -->\\r\\n    </div>\\r\\n  </div>\\r\\n  <div class=\\\"success-rate-container\\\">\\r\\n    {#if successRate != undefined}\\r\\n      <div class=\\\"success-rate-title\\\">Success rate</div>\\r\\n      <div class=\\\"errors\\\">\\r\\n        {#each successRate as value, i}\\r\\n          <div\\r\\n            class=\\\"error\\\"\\r\\n            style=\\\"background: {colors[Math.floor(value * 10) + 1]}\\\"\\r\\n            title=\\\"{(value * 100).toFixed(1)}%\\\"\\r\\n          />\\r\\n        {/each}\\r\\n      </div>\\r\\n    {/if}\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    margin: 0;\\r\\n    width: 100%;\\r\\n  }\\r\\n  .errors {\\r\\n    display: flex;\\r\\n    margin-top: 8px;\\r\\n  }\\r\\n  .error {\\r\\n    background: var(--highlight);\\r\\n    flex: 1;\\r\\n    height: 40px;\\r\\n    margin: 0 1px;\\r\\n    border-radius: 1px;\\r\\n  }\\r\\n  .success-rate-container {\\r\\n    text-align: left;\\r\\n    font-size: 0.9em;\\r\\n    color: #707070;\\r\\n    /* color: rgb(68, 68, 68); */\\r\\n  }\\r\\n  .success-rate-container {\\r\\n    margin: 1.5em 2em 2em;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAkPE,KAAK,cAAC,CAAC,AACL,MAAM,CAAE,CAAC,CACT,KAAK,CAAE,IAAI,AACb,CAAC,AACD,OAAO,cAAC,CAAC,AACP,OAAO,CAAE,IAAI,CACb,UAAU,CAAE,GAAG,AACjB,CAAC,AACD,MAAM,cAAC,CAAC,AACN,UAAU,CAAE,IAAI,WAAW,CAAC,CAC5B,IAAI,CAAE,CAAC,CACP,MAAM,CAAE,IAAI,CACZ,MAAM,CAAE,CAAC,CAAC,GAAG,CACb,aAAa,CAAE,GAAG,AACpB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,UAAU,CAAE,IAAI,CAChB,SAAS,CAAE,KAAK,CAChB,KAAK,CAAE,OAAO,AAEhB,CAAC,AACD,uBAAuB,cAAC,CAAC,AACvB,MAAM,CAAE,KAAK,CAAC,GAAG,CAAC,GAAG,AACvB,CAAC\"}"
};

function pastMonth(date) {
	let monthAgo = new Date();
	monthAgo.setDate(monthAgo.getDate() - 30);
	return date > monthAgo;
}

function daysAgo(date) {
	let now = new Date();
	return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));
}

function responseTimePlotLayout() {
	let monthAgo = new Date();
	monthAgo.setDate(monthAgo.getDate() - 30);
	let tomorrow = new Date();
	tomorrow.setDate(tomorrow.getDate() + 1);

	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
		hovermode: "closest",
		plot_bgcolor: "transparent",
		paper_bgcolor: "transparent",
		height: 150,
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
			range: [monthAgo, tomorrow],
			visible: false
		},
		dragmode: false
	};
}

function requestsFreqPlotLayout() {
	let monthAgo = new Date();
	monthAgo.setDate(monthAgo.getDate() - 30);
	let tomorrow = new Date();
	tomorrow.setDate(tomorrow.getDate() + 1);

	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
		hovermode: "closest",
		plot_bgcolor: "transparent",
		paper_bgcolor: "transparent",
		height: 150,
		yaxis: {
			title: { text: "Requests" },
			gridcolor: "gray",
			showgrid: false,
			fixedrange: true
		},
		xaxis: {
			title: { text: "Date" },
			fixedrange: true,
			range: [monthAgo, tomorrow],
			visible: false
		},
		dragmode: false
	};
}

const PastMonth = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let colors = [
		"#444444",
		"#E46161",
		"#F18359",
		"#F5A65A",
		"#F3C966",
		"#EBEB81",
		"#C7E57D",
		"#A1DF7E",
		"#77D884",
		"#3FCF8E"
	]; // Green

	function setSuccessRate() {
		let success = {};

		for (let i = 0; i < data.length; i++) {
			let date = new Date(data[i].created_at);

			if (pastMonth(date)) {
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
			}
		}

		let successArr = new Array(60).fill(-0.1); // -0.1 -> 0

		for (let date in success) {
			let idx = daysAgo(new Date(date));
			successArr[successArr.length - idx] = success[date].successful / success[date].total;
		}

		successRate = successArr;
	}

	function responseTimeLine() {
		let points = [];

		for (let i = 0; i < data.length; i++) {
			points.push([new Date(data[i].created_at), data[i].response_time]);
		}

		points.sort((a, b) => {
			return a[0] - b[0];
		});

		let dates = [];
		let responses = [];

		for (let i = 0; i < points.length; i++) {
			dates.push(points[i][0]);
			responses.push(points[i][1]);
		}

		return [
			{
				x: dates,
				y: responses,
				mode: "lines",
				line: { color: "#3fcf8e" },
				hovertemplate: `<b>%{y:.1f}ms avg</b><br>%{x|%d %b %Y}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	function responseTimePlotData() {
		return {
			data: responseTimeLine(),
			layout: responseTimePlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genResponseTimePlot() {
		let plotData = responseTimePlotData();

		//@ts-ignore
		new Plotly.newPlot(responseTimePlotDiv, plotData.data, plotData.layout, plotData.config);
	}

	function requestsFreqLine() {
		let requestFreq = {};

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

	function requestsFreqPlotData() {
		return {
			data: requestsFreqLine(),
			layout: requestsFreqPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genRequestsFreqPlot() {
		let plotData = requestsFreqPlotData();

		//@ts-ignore
		new Plotly.newPlot(requestsFreqPlotDiv, plotData.data, plotData.layout, plotData.config);
	}

	function build() {
		setSuccessRate();
		genResponseTimePlot();
		genRequestsFreqPlot();
	}

	let successRate;
	let responseTimePlotDiv;
	let requestsFreqPlotDiv;

	onMount(() => {
		build();
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$3);

	return `<div class="${"card svelte-ik169t"}"><div class="${"card-title"}">Past Month</div>
  <div id="${"plotly"}"><div id="${"plotDiv"}"${add_attribute("this", requestsFreqPlotDiv, 0)}></div></div>
  <div id="${"plotly"}"><div id="${"plotDiv"}"${add_attribute("this", responseTimePlotDiv, 0)}></div></div>
  <div class="${"success-rate-container svelte-ik169t"}">${successRate != undefined
	? `<div class="${"success-rate-title"}">Success rate</div>
      <div class="${"errors svelte-ik169t"}">${each(successRate, (value, i) => {
			return `<div class="${"error svelte-ik169t"}" style="${"background: " + escape(colors[Math.floor(value * 10) + 1], true)}" title="${escape((value * 100).toFixed(1), true) + "%"}"></div>`;
		})}</div>`
	: ``}</div>
</div>`;
});

/* src\components\Browser.svelte generated by Svelte v3.53.1 */

const css$2 = {
	code: ".card.svelte-1k7nusv{margin:2em 2em 2em 0;padding-bottom:1em}#plotDiv.svelte-1k7nusv{margin:0 20px}",
	map: "{\"version\":3,\"file\":\"Browser.svelte\",\"sources\":[\"Browser.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction thisWeek(date) {\\r\\n    let weekAgo = new Date();\\r\\n    weekAgo.setDate(weekAgo.getDate() - 7);\\r\\n    return date > weekAgo;\\r\\n}\\r\\nfunction getBrowser(userAgent) {\\r\\n    if (userAgent.match(/Seamonkey\\\\//)) {\\r\\n        return 'Seamonkey';\\r\\n    }\\r\\n    else if (userAgent.match(/Firefox\\\\//)) {\\r\\n        return 'Firefox';\\r\\n    }\\r\\n    else if (userAgent.match(/Chrome\\\\//)) {\\r\\n        return 'Chrome';\\r\\n    }\\r\\n    else if (userAgent.match(/Chromium\\\\//)) {\\r\\n        return 'Chromium';\\r\\n    }\\r\\n    else if (userAgent.match(/Safari\\\\//)) {\\r\\n        return 'Safari';\\r\\n    }\\r\\n    else if (userAgent.match(/Edg\\\\//)) {\\r\\n        return 'Edge';\\r\\n    }\\r\\n    else if (userAgent.match(/OPR\\\\//) || userAgent.match(/Opera\\\\//)) {\\r\\n        return 'Opera';\\r\\n    }\\r\\n    else if (userAgent.match(/; MSIE /) || userAgent.match(/Trident\\\\//)) {\\r\\n        return 'Internet Explorer';\\r\\n    }\\r\\n    else if (userAgent.match(/curl\\\\//)) {\\r\\n        return 'Curl';\\r\\n    }\\r\\n    else if (userAgent.match(/PostmanRuntime\\\\//)) {\\r\\n        return 'Postman';\\r\\n    }\\r\\n    else if (userAgent.match(/insomnia\\\\//)) {\\r\\n        return 'Insomnia';\\r\\n    }\\r\\n    else {\\r\\n        return 'Other';\\r\\n    }\\r\\n}\\r\\nfunction browserPlotLayout() {\\r\\n    let monthAgo = new Date();\\r\\n    monthAgo.setDate(monthAgo.getDate() - 30);\\r\\n    let tomorrow = new Date();\\r\\n    tomorrow.setDate(tomorrow.getDate() + 1);\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 180,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Requests\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nlet colors = [\\r\\n    \\\"#3FCF8E\\\",\\r\\n    \\\"#E46161\\\",\\r\\n    \\\"#EBEB81\\\", // Yellow\\r\\n];\\r\\nfunction pieChart() {\\r\\n    let browserCount = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let browser = getBrowser(data[i].user_agent);\\r\\n        if (!(browser in browserCount)) {\\r\\n            browserCount[browser] = 0;\\r\\n        }\\r\\n        browserCount[browser]++;\\r\\n    }\\r\\n    let browsers = [];\\r\\n    let count = [];\\r\\n    for (let browser in browserCount) {\\r\\n        browsers.push(browser);\\r\\n        count.push(browserCount[browser]);\\r\\n    }\\r\\n    return [{\\r\\n            values: count,\\r\\n            labels: browsers,\\r\\n            type: 'pie',\\r\\n            marker: {\\r\\n                colors: colors\\r\\n            },\\r\\n        }];\\r\\n}\\r\\nfunction browserPlotData() {\\r\\n    return {\\r\\n        data: pieChart(),\\r\\n        layout: browserPlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genPlot() {\\r\\n    let plotData = browserPlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\n//   function build() {\\r\\n//     let totalRequests = 0;\\r\\n//     for (let i = 0; i < data.length; i++) {\\r\\n//       let date = new Date(data[i].created_at);\\r\\n//       if (thisWeek(date)) {\\r\\n//         totalRequests++;\\r\\n//       }\\r\\n//     }\\r\\n//     requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);\\r\\n//   }\\r\\nlet plotDiv;\\r\\nonMount(() => {\\r\\n    genPlot();\\r\\n});\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Last week\\\">\\r\\n  <div class=\\\"card-title\\\">Browser</div>\\r\\n  <div id=\\\"plotly\\\">\\r\\n    <div id=\\\"plotDiv\\\" bind:this={plotDiv}>\\r\\n      <!-- Plotly chart will be drawn inside this DIV -->\\r\\n    </div>\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    margin: 2em 2em 2em 0;\\r\\n    padding-bottom: 1em;\\r\\n  }\\r\\n  #plotDiv {\\r\\n    margin: 0 20px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AA6IE,KAAK,eAAC,CAAC,AACL,MAAM,CAAE,GAAG,CAAC,GAAG,CAAC,GAAG,CAAC,CAAC,CACrB,cAAc,CAAE,GAAG,AACrB,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,MAAM,CAAE,CAAC,CAAC,IAAI,AAChB,CAAC\"}"
};

function getBrowser(userAgent) {
	if (userAgent.match(/Seamonkey\//)) {
		return 'Seamonkey';
	} else if (userAgent.match(/Firefox\//)) {
		return 'Firefox';
	} else if (userAgent.match(/Chrome\//)) {
		return 'Chrome';
	} else if (userAgent.match(/Chromium\//)) {
		return 'Chromium';
	} else if (userAgent.match(/Safari\//)) {
		return 'Safari';
	} else if (userAgent.match(/Edg\//)) {
		return 'Edge';
	} else if (userAgent.match(/OPR\//) || userAgent.match(/Opera\//)) {
		return 'Opera';
	} else if (userAgent.match(/; MSIE /) || userAgent.match(/Trident\//)) {
		return 'Internet Explorer';
	} else if (userAgent.match(/curl\//)) {
		return 'Curl';
	} else if (userAgent.match(/PostmanRuntime\//)) {
		return 'Postman';
	} else if (userAgent.match(/insomnia\//)) {
		return 'Insomnia';
	} else {
		return 'Other';
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

const Browser = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let colors = ["#3FCF8E", "#E46161", "#EBEB81"]; // Yellow

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
				type: 'pie',
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

	//   function build() {
	//     let totalRequests = 0;
	//     for (let i = 0; i < data.length; i++) {
	//       let date = new Date(data[i].created_at);
	//       if (thisWeek(date)) {
	//         totalRequests++;
	//       }
	//     }
	//     requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);
	//   }
	let plotDiv;

	onMount(() => {
		genPlot();
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$2);

	return `<div class="${"card svelte-1k7nusv"}" title="${"Last week"}"><div class="${"card-title"}">Browser</div>
  <div id="${"plotly"}"><div id="${"plotDiv"}" class="${"svelte-1k7nusv"}"${add_attribute("this", plotDiv, 0)}></div></div>
</div>`;
});

/* src\components\OperatingSystem.svelte generated by Svelte v3.53.1 */

const css$1 = {
	code: ".card.svelte-1k7nusv{margin:2em 2em 2em 0;padding-bottom:1em}#plotDiv.svelte-1k7nusv{margin:0 20px}",
	map: "{\"version\":3,\"file\":\"OperatingSystem.svelte\",\"sources\":[\"OperatingSystem.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nfunction thisWeek(date) {\\r\\n    let weekAgo = new Date();\\r\\n    weekAgo.setDate(weekAgo.getDate() - 7);\\r\\n    return date > weekAgo;\\r\\n}\\r\\nfunction getOS(userAgent) {\\r\\n    if (userAgent.match(/Win16/)) {\\r\\n        return 'Windows 3.11';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows 95)|(Win95)|(Windows_95)/)) {\\r\\n        return 'Windows 95';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows 98)|(Win98)/)) {\\r\\n        return 'Windows 98';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 5.0)|(Windows 2000)/)) {\\r\\n        return 'Windows 2000';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 5.1)|(Windows XP)/)) {\\r\\n        return 'Windows XP';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 5.2)/)) {\\r\\n        return 'Windows Server 2003';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 6.0)/)) {\\r\\n        return 'Windows Vista';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 6.1)/)) {\\r\\n        return 'Windows 7';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 6.2)/)) {\\r\\n        return 'Windows 8';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 10.0)/)) {\\r\\n        return 'Windows 10';\\r\\n    }\\r\\n    else if (userAgent.match(/(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/)) {\\r\\n        return 'Windows NT 4.0';\\r\\n    }\\r\\n    else if (userAgent.match(/Windows ME/)) {\\r\\n        return 'Windows ME';\\r\\n    }\\r\\n    else if (userAgent.match(/OpenBSD/)) {\\r\\n        return 'OpenBSE';\\r\\n    }\\r\\n    else if (userAgent.match(/SunOS/)) {\\r\\n        return 'SunOS';\\r\\n    }\\r\\n    else if (userAgent.match(/(Linux)|(X11)/)) {\\r\\n        return 'Linux';\\r\\n    }\\r\\n    else if (userAgent.match(/(Mac_PowerPC)|(Macintosh)/)) {\\r\\n        return 'MacOS';\\r\\n    }\\r\\n    else if (userAgent.match(/QNX/)) {\\r\\n        return 'QNX';\\r\\n    }\\r\\n    else if (userAgent.match(/BeOS/)) {\\r\\n        return 'BeOS';\\r\\n    }\\r\\n    else if (userAgent.match(/OS\\\\/2/)) {\\r\\n        return 'OS/2';\\r\\n    }\\r\\n    else if (userAgent.match(/(nuhk)|(Googlebot)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\\\\/Teoma)|(ia_archiver)/)) {\\r\\n        return 'Search Bot';\\r\\n    }\\r\\n    else {\\r\\n        return 'Unknown';\\r\\n    }\\r\\n}\\r\\nfunction osPlotLayout() {\\r\\n    let monthAgo = new Date();\\r\\n    monthAgo.setDate(monthAgo.getDate() - 30);\\r\\n    let tomorrow = new Date();\\r\\n    tomorrow.setDate(tomorrow.getDate() + 1);\\r\\n    return {\\r\\n        title: false,\\r\\n        autosize: true,\\r\\n        margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },\\r\\n        hovermode: \\\"closest\\\",\\r\\n        plot_bgcolor: \\\"transparent\\\",\\r\\n        paper_bgcolor: \\\"transparent\\\",\\r\\n        height: 180,\\r\\n        yaxis: {\\r\\n            title: { text: \\\"Requests\\\" },\\r\\n            gridcolor: \\\"gray\\\",\\r\\n            showgrid: false,\\r\\n            fixedrange: true,\\r\\n        },\\r\\n        xaxis: {\\r\\n            visible: false,\\r\\n        },\\r\\n        dragmode: false,\\r\\n    };\\r\\n}\\r\\nlet colors = [\\r\\n    \\\"#3FCF8E\\\",\\r\\n    \\\"#E46161\\\",\\r\\n    \\\"#EBEB81\\\", // Yellow\\r\\n];\\r\\nfunction pieChart() {\\r\\n    let osCount = {};\\r\\n    for (let i = 0; i < data.length; i++) {\\r\\n        let os = getOS(data[i].user_agent);\\r\\n        if (!(os in osCount)) {\\r\\n            osCount[os] = 0;\\r\\n        }\\r\\n        osCount[os]++;\\r\\n    }\\r\\n    let os = [];\\r\\n    let count = [];\\r\\n    for (let browser in osCount) {\\r\\n        os.push(browser);\\r\\n        count.push(osCount[browser]);\\r\\n    }\\r\\n    return [{\\r\\n            values: count,\\r\\n            labels: os,\\r\\n            type: 'pie',\\r\\n            marker: {\\r\\n                colors: colors\\r\\n            },\\r\\n        }];\\r\\n}\\r\\nfunction osPlotData() {\\r\\n    return {\\r\\n        data: pieChart(),\\r\\n        layout: osPlotLayout(),\\r\\n        config: {\\r\\n            responsive: true,\\r\\n            showSendToCloud: false,\\r\\n            displayModeBar: false,\\r\\n        },\\r\\n    };\\r\\n}\\r\\nfunction genPlot() {\\r\\n    let plotData = osPlotData();\\r\\n    //@ts-ignore\\r\\n    new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);\\r\\n}\\r\\n// function build() {\\r\\n//   let totalRequests = 0;\\r\\n//   for (let i = 0; i < data.length; i++) {\\r\\n//     let date = new Date(data[i].created_at);\\r\\n//     if (thisWeek(date)) {\\r\\n//       totalRequests++;\\r\\n//     }\\r\\n//   }\\r\\n//   requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);\\r\\n// }\\r\\nlet plotDiv;\\r\\nonMount(() => {\\r\\n    genPlot();\\r\\n});\\r\\nexport let data;\\r\\n</script>\\r\\n\\r\\n<div class=\\\"card\\\" title=\\\"Last week\\\">\\r\\n  <div class=\\\"card-title\\\">OS</div>\\r\\n  <div id=\\\"plotly\\\">\\r\\n    <div id=\\\"plotDiv\\\" bind:this={plotDiv}>\\r\\n      <!-- Plotly chart will be drawn inside this DIV -->\\r\\n    </div>\\r\\n  </div>\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .card {\\r\\n    margin: 2em 2em 2em 0;\\r\\n    padding-bottom: 1em;\\r\\n  }\\r\\n  #plotDiv {\\r\\n    margin: 0 20px;\\r\\n  }\\r\\n\\r\\n</style>\\r\\n\"],\"names\":[],\"mappings\":\"AAwKE,KAAK,eAAC,CAAC,AACL,MAAM,CAAE,GAAG,CAAC,GAAG,CAAC,GAAG,CAAC,CAAC,CACrB,cAAc,CAAE,GAAG,AACrB,CAAC,AACD,QAAQ,eAAC,CAAC,AACR,MAAM,CAAE,CAAC,CAAC,IAAI,AAChB,CAAC\"}"
};

function getOS(userAgent) {
	if (userAgent.match(/Win16/)) {
		return 'Windows 3.11';
	} else if (userAgent.match(/(Windows 95)|(Win95)|(Windows_95)/)) {
		return 'Windows 95';
	} else if (userAgent.match(/(Windows 98)|(Win98)/)) {
		return 'Windows 98';
	} else if (userAgent.match(/(Windows NT 5.0)|(Windows 2000)/)) {
		return 'Windows 2000';
	} else if (userAgent.match(/(Windows NT 5.1)|(Windows XP)/)) {
		return 'Windows XP';
	} else if (userAgent.match(/(Windows NT 5.2)/)) {
		return 'Windows Server 2003';
	} else if (userAgent.match(/(Windows NT 6.0)/)) {
		return 'Windows Vista';
	} else if (userAgent.match(/(Windows NT 6.1)/)) {
		return 'Windows 7';
	} else if (userAgent.match(/(Windows NT 6.2)/)) {
		return 'Windows 8';
	} else if (userAgent.match(/(Windows NT 10.0)/)) {
		return 'Windows 10';
	} else if (userAgent.match(/(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/)) {
		return 'Windows NT 4.0';
	} else if (userAgent.match(/Windows ME/)) {
		return 'Windows ME';
	} else if (userAgent.match(/OpenBSD/)) {
		return 'OpenBSE';
	} else if (userAgent.match(/SunOS/)) {
		return 'SunOS';
	} else if (userAgent.match(/(Linux)|(X11)/)) {
		return 'Linux';
	} else if (userAgent.match(/(Mac_PowerPC)|(Macintosh)/)) {
		return 'MacOS';
	} else if (userAgent.match(/QNX/)) {
		return 'QNX';
	} else if (userAgent.match(/BeOS/)) {
		return 'BeOS';
	} else if (userAgent.match(/OS\/2/)) {
		return 'OS/2';
	} else if (userAgent.match(/(nuhk)|(Googlebot)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\/Teoma)|(ia_archiver)/)) {
		return 'Search Bot';
	} else {
		return 'Unknown';
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

const OperatingSystem = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let colors = ["#3FCF8E", "#E46161", "#EBEB81"]; // Yellow

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
				type: 'pie',
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

	// function build() {
	//   let totalRequests = 0;
	//   for (let i = 0; i < data.length; i++) {
	//     let date = new Date(data[i].created_at);
	//     if (thisWeek(date)) {
	//       totalRequests++;
	//     }
	//   }
	//   requestsPerHour = ((24 * 7) / totalRequests).toFixed(2);
	// }
	let plotDiv;

	onMount(() => {
		genPlot();
	});

	let { data } = $$props;
	if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
	$$result.css.add(css$1);

	return `<div class="${"card svelte-1k7nusv"}" title="${"Last week"}"><div class="${"card-title"}">OS</div>
  <div id="${"plotly"}"><div id="${"plotDiv"}" class="${"svelte-1k7nusv"}"${add_attribute("this", plotDiv, 0)}></div></div>
</div>`;
});

/* src\routes\Dashboard.svelte generated by Svelte v3.53.1 */

const css = {
	code: ".dashboard.svelte-95rck4{margin:5em;display:flex}.row.svelte-95rck4{display:flex}.right.svelte-95rck4{flex-grow:1}",
	map: "{\"version\":3,\"file\":\"Dashboard.svelte\",\"sources\":[\"Dashboard.svelte\"],\"sourcesContent\":[\"<script lang=\\\"ts\\\">import { onMount } from \\\"svelte\\\";\\r\\nimport Requests from \\\"../components/Requests.svelte\\\";\\r\\nimport ResponseTimes from \\\"../components/ResponseTimes.svelte\\\";\\r\\nimport Endpoints from \\\"../components/Endpoints.svelte\\\";\\r\\nimport Footer from \\\"../components/Footer.svelte\\\";\\r\\nimport SuccessRate from \\\"../components/SuccessRate.svelte\\\";\\r\\nimport PastMonth from \\\"../components/PastMonth.svelte\\\";\\r\\nimport Browser from \\\"../components/Browser.svelte\\\";\\r\\nimport OperatingSystem from \\\"../components/OperatingSystem.svelte\\\";\\r\\nfunction formatUUID(userID) {\\r\\n    return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(12, 16)}-${userID.slice(16, 20)}-${userID.slice(20)}`;\\r\\n}\\r\\nasync function fetchData() {\\r\\n    userID = formatUUID(userID);\\r\\n    // Fetch page ID\\r\\n    const response = await fetch(`https://api-analytics-server.vercel.app/api/data/${userID}`);\\r\\n    if (response.status == 200) {\\r\\n        const json = await response.json();\\r\\n        data = json.value;\\r\\n        console.log(data);\\r\\n    }\\r\\n}\\r\\nlet data;\\r\\nonMount(() => {\\r\\n    fetchData();\\r\\n});\\r\\nexport let userID;\\r\\n</script>\\r\\n\\r\\n<div>\\r\\n  <!-- <h1>Dashboard</h1> -->\\r\\n  {#if data != undefined}\\r\\n    <div class=\\\"dashboard\\\">\\r\\n      <div class=\\\"left\\\">\\r\\n        <div class=\\\"row\\\">\\r\\n          <Requests {data} />\\r\\n          <SuccessRate {data} />\\r\\n        </div>\\r\\n        <ResponseTimes {data} />\\r\\n        <Endpoints {data} />\\r\\n      </div>\\r\\n      <div class=\\\"right\\\">\\r\\n        <PastMonth {data} />\\r\\n        <div class=\\\"row\\\">\\r\\n\\r\\n          <OperatingSystem {data} />\\r\\n          <Browser {data} />\\r\\n        </div>\\r\\n      </div>\\r\\n    </div>\\r\\n    {/if}\\r\\n  <Footer />\\r\\n</div>\\r\\n\\r\\n<style>\\r\\n  .dashboard {\\r\\n    margin: 5em;\\r\\n    display: flex;\\r\\n  }\\r\\n  .row {\\r\\n    display: flex;\\r\\n  }\\r\\n  .right {\\r\\n    flex-grow: 1;\\r\\n  }\\r\\n\\r\\n</style>\"],\"names\":[],\"mappings\":\"AAuDE,UAAU,cAAC,CAAC,AACV,MAAM,CAAE,GAAG,CACX,OAAO,CAAE,IAAI,AACf,CAAC,AACD,IAAI,cAAC,CAAC,AACJ,OAAO,CAAE,IAAI,AACf,CAAC,AACD,MAAM,cAAC,CAAC,AACN,SAAS,CAAE,CAAC,AACd,CAAC\"}"
};

function formatUUID(userID) {
	return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(12, 16)}-${userID.slice(16, 20)}-${userID.slice(20)}`;
}

const Dashboard = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	async function fetchData() {
		userID = formatUUID(userID);

		// Fetch page ID
		const response = await fetch(`https://api-analytics-server.vercel.app/api/data/${userID}`);

		if (response.status == 200) {
			const json = await response.json();
			data = json.value;
			console.log(data);
		}
	}

	let data;

	onMount(() => {
		fetchData();
	});

	let { userID } = $$props;
	if ($$props.userID === void 0 && $$bindings.userID && userID !== void 0) $$bindings.userID(userID);
	$$result.css.add(css);

	return `<div>
  ${data != undefined
	? `<div class="${"dashboard svelte-95rck4"}"><div class="${"left"}"><div class="${"row svelte-95rck4"}">${validate_component(Requests, "Requests").$$render($$result, { data }, {}, {})}
          ${validate_component(SuccessRate, "SuccessRate").$$render($$result, { data }, {}, {})}</div>
        ${validate_component(ResponseTimes, "ResponseTimes").$$render($$result, { data }, {}, {})}
        ${validate_component(Endpoints, "Endpoints").$$render($$result, { data }, {}, {})}</div>
      <div class="${"right svelte-95rck4"}">${validate_component(PastMonth, "PastMonth").$$render($$result, { data }, {}, {})}
        <div class="${"row svelte-95rck4"}">${validate_component(OperatingSystem, "OperatingSystem").$$render($$result, { data }, {}, {})}
          ${validate_component(Browser, "Browser").$$render($$result, { data }, {}, {})}</div></div></div>`
	: ``}
  ${validate_component(Footer, "Footer").$$render($$result, {}, {}, {})}
</div>`;
});

/* src\App.svelte generated by Svelte v3.53.1 */

const App = create_ssr_component(($$result, $$props, $$bindings, slots) => {
	let { url = "" } = $$props;
	if ($$props.url === void 0 && $$bindings.url && url !== void 0) $$bindings.url(url);

	return `${validate_component(Router, "Router").$$render($$result, { url }, {}, {
		default: () => {
			return `
    ${validate_component(Route, "Route").$$render($$result, { path: "/generate", component: Generate }, {}, {})}
    ${validate_component(Route, "Route").$$render($$result, { path: "/dashboard", component: SignIn }, {}, {})}
    ${validate_component(Route, "Route").$$render(
				$$result,
				{
					path: "/dashboard/:userID",
					component: Dashboard
				},
				{},
				{}
			)}`;
		}
	})}`;
});

module.exports = App;
