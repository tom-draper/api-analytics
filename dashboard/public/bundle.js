
(function(l, r) { if (!l || l.getElementById('livereloadscript')) return; r = l.createElement('script'); r.async = 1; r.src = '//' + (self.location.host || 'localhost').split(':')[0] + ':35729/livereload.js?snipver=1'; r.id = 'livereloadscript'; l.getElementsByTagName('head')[0].appendChild(r) })(self.document);
(function () {
    'use strict';

    function noop() { }
    function assign(tar, src) {
        // @ts-ignore
        for (const k in src)
            tar[k] = src[k];
        return tar;
    }
    function add_location(element, file, line, column, char) {
        element.__svelte_meta = {
            loc: { file, line, column, char }
        };
    }
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
    let src_url_equal_anchor;
    function src_url_equal(element_src, url) {
        if (!src_url_equal_anchor) {
            src_url_equal_anchor = document.createElement('a');
        }
        src_url_equal_anchor.href = url;
        return element_src === src_url_equal_anchor.href;
    }
    function is_empty(obj) {
        return Object.keys(obj).length === 0;
    }
    function validate_store(store, name) {
        if (store != null && typeof store.subscribe !== 'function') {
            throw new Error(`'${name}' is not a store with a 'subscribe' method`);
        }
    }
    function subscribe(store, ...callbacks) {
        if (store == null) {
            return noop;
        }
        const unsub = store.subscribe(...callbacks);
        return unsub.unsubscribe ? () => unsub.unsubscribe() : unsub;
    }
    function component_subscribe(component, store, callback) {
        component.$$.on_destroy.push(subscribe(store, callback));
    }
    function create_slot(definition, ctx, $$scope, fn) {
        if (definition) {
            const slot_ctx = get_slot_context(definition, ctx, $$scope, fn);
            return definition[0](slot_ctx);
        }
    }
    function get_slot_context(definition, ctx, $$scope, fn) {
        return definition[1] && fn
            ? assign($$scope.ctx.slice(), definition[1](fn(ctx)))
            : $$scope.ctx;
    }
    function get_slot_changes(definition, $$scope, dirty, fn) {
        if (definition[2] && fn) {
            const lets = definition[2](fn(dirty));
            if ($$scope.dirty === undefined) {
                return lets;
            }
            if (typeof lets === 'object') {
                const merged = [];
                const len = Math.max($$scope.dirty.length, lets.length);
                for (let i = 0; i < len; i += 1) {
                    merged[i] = $$scope.dirty[i] | lets[i];
                }
                return merged;
            }
            return $$scope.dirty | lets;
        }
        return $$scope.dirty;
    }
    function update_slot_base(slot, slot_definition, ctx, $$scope, slot_changes, get_slot_context_fn) {
        if (slot_changes) {
            const slot_context = get_slot_context(slot_definition, ctx, $$scope, get_slot_context_fn);
            slot.p(slot_context, slot_changes);
        }
    }
    function get_all_dirty_from_scope($$scope) {
        if ($$scope.ctx.length > 32) {
            const dirty = [];
            const length = $$scope.ctx.length / 32;
            for (let i = 0; i < length; i++) {
                dirty[i] = -1;
            }
            return dirty;
        }
        return -1;
    }
    function exclude_internal_props(props) {
        const result = {};
        for (const k in props)
            if (k[0] !== '$')
                result[k] = props[k];
        return result;
    }
    function null_to_empty(value) {
        return value == null ? '' : value;
    }

    // Track which nodes are claimed during hydration. Unclaimed nodes can then be removed from the DOM
    // at the end of hydration without touching the remaining nodes.
    let is_hydrating = false;
    function start_hydrating() {
        is_hydrating = true;
    }
    function end_hydrating() {
        is_hydrating = false;
    }
    function upper_bound(low, high, key, value) {
        // Return first index of value larger than input value in the range [low, high)
        while (low < high) {
            const mid = low + ((high - low) >> 1);
            if (key(mid) <= value) {
                low = mid + 1;
            }
            else {
                high = mid;
            }
        }
        return low;
    }
    function init_hydrate(target) {
        if (target.hydrate_init)
            return;
        target.hydrate_init = true;
        // We know that all children have claim_order values since the unclaimed have been detached if target is not <head>
        let children = target.childNodes;
        // If target is <head>, there may be children without claim_order
        if (target.nodeName === 'HEAD') {
            const myChildren = [];
            for (let i = 0; i < children.length; i++) {
                const node = children[i];
                if (node.claim_order !== undefined) {
                    myChildren.push(node);
                }
            }
            children = myChildren;
        }
        /*
        * Reorder claimed children optimally.
        * We can reorder claimed children optimally by finding the longest subsequence of
        * nodes that are already claimed in order and only moving the rest. The longest
        * subsequence of nodes that are claimed in order can be found by
        * computing the longest increasing subsequence of .claim_order values.
        *
        * This algorithm is optimal in generating the least amount of reorder operations
        * possible.
        *
        * Proof:
        * We know that, given a set of reordering operations, the nodes that do not move
        * always form an increasing subsequence, since they do not move among each other
        * meaning that they must be already ordered among each other. Thus, the maximal
        * set of nodes that do not move form a longest increasing subsequence.
        */
        // Compute longest increasing subsequence
        // m: subsequence length j => index k of smallest value that ends an increasing subsequence of length j
        const m = new Int32Array(children.length + 1);
        // Predecessor indices + 1
        const p = new Int32Array(children.length);
        m[0] = -1;
        let longest = 0;
        for (let i = 0; i < children.length; i++) {
            const current = children[i].claim_order;
            // Find the largest subsequence length such that it ends in a value less than our current value
            // upper_bound returns first greater value, so we subtract one
            // with fast path for when we are on the current longest subsequence
            const seqLen = ((longest > 0 && children[m[longest]].claim_order <= current) ? longest + 1 : upper_bound(1, longest, idx => children[m[idx]].claim_order, current)) - 1;
            p[i] = m[seqLen] + 1;
            const newLen = seqLen + 1;
            // We can guarantee that current is the smallest value. Otherwise, we would have generated a longer sequence.
            m[newLen] = i;
            longest = Math.max(newLen, longest);
        }
        // The longest increasing subsequence of nodes (initially reversed)
        const lis = [];
        // The rest of the nodes, nodes that will be moved
        const toMove = [];
        let last = children.length - 1;
        for (let cur = m[longest] + 1; cur != 0; cur = p[cur - 1]) {
            lis.push(children[cur - 1]);
            for (; last >= cur; last--) {
                toMove.push(children[last]);
            }
            last--;
        }
        for (; last >= 0; last--) {
            toMove.push(children[last]);
        }
        lis.reverse();
        // We sort the nodes being moved to guarantee that their insertion order matches the claim order
        toMove.sort((a, b) => a.claim_order - b.claim_order);
        // Finally, we move the nodes
        for (let i = 0, j = 0; i < toMove.length; i++) {
            while (j < lis.length && toMove[i].claim_order >= lis[j].claim_order) {
                j++;
            }
            const anchor = j < lis.length ? lis[j] : null;
            target.insertBefore(toMove[i], anchor);
        }
    }
    function append_hydration(target, node) {
        if (is_hydrating) {
            init_hydrate(target);
            if ((target.actual_end_child === undefined) || ((target.actual_end_child !== null) && (target.actual_end_child.parentNode !== target))) {
                target.actual_end_child = target.firstChild;
            }
            // Skip nodes of undefined ordering
            while ((target.actual_end_child !== null) && (target.actual_end_child.claim_order === undefined)) {
                target.actual_end_child = target.actual_end_child.nextSibling;
            }
            if (node !== target.actual_end_child) {
                // We only insert if the ordering of this node should be modified or the parent node is not target
                if (node.claim_order !== undefined || node.parentNode !== target) {
                    target.insertBefore(node, target.actual_end_child);
                }
            }
            else {
                target.actual_end_child = node.nextSibling;
            }
        }
        else if (node.parentNode !== target || node.nextSibling !== null) {
            target.appendChild(node);
        }
    }
    function insert_hydration(target, node, anchor) {
        if (is_hydrating && !anchor) {
            append_hydration(target, node);
        }
        else if (node.parentNode !== target || node.nextSibling != anchor) {
            target.insertBefore(node, anchor || null);
        }
    }
    function detach(node) {
        if (node.parentNode) {
            node.parentNode.removeChild(node);
        }
    }
    function destroy_each(iterations, detaching) {
        for (let i = 0; i < iterations.length; i += 1) {
            if (iterations[i])
                iterations[i].d(detaching);
        }
    }
    function element(name) {
        return document.createElement(name);
    }
    function text(data) {
        return document.createTextNode(data);
    }
    function space() {
        return text(' ');
    }
    function empty() {
        return text('');
    }
    function listen(node, event, handler, options) {
        node.addEventListener(event, handler, options);
        return () => node.removeEventListener(event, handler, options);
    }
    function attr(node, attribute, value) {
        if (value == null)
            node.removeAttribute(attribute);
        else if (node.getAttribute(attribute) !== value)
            node.setAttribute(attribute, value);
    }
    function children(element) {
        return Array.from(element.childNodes);
    }
    function init_claim_info(nodes) {
        if (nodes.claim_info === undefined) {
            nodes.claim_info = { last_index: 0, total_claimed: 0 };
        }
    }
    function claim_node(nodes, predicate, processNode, createNode, dontUpdateLastIndex = false) {
        // Try to find nodes in an order such that we lengthen the longest increasing subsequence
        init_claim_info(nodes);
        const resultNode = (() => {
            // We first try to find an element after the previous one
            for (let i = nodes.claim_info.last_index; i < nodes.length; i++) {
                const node = nodes[i];
                if (predicate(node)) {
                    const replacement = processNode(node);
                    if (replacement === undefined) {
                        nodes.splice(i, 1);
                    }
                    else {
                        nodes[i] = replacement;
                    }
                    if (!dontUpdateLastIndex) {
                        nodes.claim_info.last_index = i;
                    }
                    return node;
                }
            }
            // Otherwise, we try to find one before
            // We iterate in reverse so that we don't go too far back
            for (let i = nodes.claim_info.last_index - 1; i >= 0; i--) {
                const node = nodes[i];
                if (predicate(node)) {
                    const replacement = processNode(node);
                    if (replacement === undefined) {
                        nodes.splice(i, 1);
                    }
                    else {
                        nodes[i] = replacement;
                    }
                    if (!dontUpdateLastIndex) {
                        nodes.claim_info.last_index = i;
                    }
                    else if (replacement === undefined) {
                        // Since we spliced before the last_index, we decrease it
                        nodes.claim_info.last_index--;
                    }
                    return node;
                }
            }
            // If we can't find any matching node, we create a new one
            return createNode();
        })();
        resultNode.claim_order = nodes.claim_info.total_claimed;
        nodes.claim_info.total_claimed += 1;
        return resultNode;
    }
    function claim_element_base(nodes, name, attributes, create_element) {
        return claim_node(nodes, (node) => node.nodeName === name, (node) => {
            const remove = [];
            for (let j = 0; j < node.attributes.length; j++) {
                const attribute = node.attributes[j];
                if (!attributes[attribute.name]) {
                    remove.push(attribute.name);
                }
            }
            remove.forEach(v => node.removeAttribute(v));
            return undefined;
        }, () => create_element(name));
    }
    function claim_element(nodes, name, attributes) {
        return claim_element_base(nodes, name, attributes, element);
    }
    function claim_text(nodes, data) {
        return claim_node(nodes, (node) => node.nodeType === 3, (node) => {
            const dataStr = '' + data;
            if (node.data.startsWith(dataStr)) {
                if (node.data.length !== dataStr.length) {
                    return node.splitText(dataStr.length);
                }
            }
            else {
                node.data = dataStr;
            }
        }, () => text(data), true // Text nodes should not update last index since it is likely not worth it to eliminate an increasing subsequence of actual elements
        );
    }
    function claim_space(nodes) {
        return claim_text(nodes, ' ');
    }
    function set_input_value(input, value) {
        input.value = value == null ? '' : value;
    }
    function set_style(node, key, value, important) {
        if (value === null) {
            node.style.removeProperty(key);
        }
        else {
            node.style.setProperty(key, value, important ? 'important' : '');
        }
    }
    function toggle_class(element, name, toggle) {
        element.classList[toggle ? 'add' : 'remove'](name);
    }
    function custom_event(type, detail, { bubbles = false, cancelable = false } = {}) {
        const e = document.createEvent('CustomEvent');
        e.initCustomEvent(type, bubbles, cancelable, detail);
        return e;
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

    const dirty_components = [];
    const binding_callbacks = [];
    const render_callbacks = [];
    const flush_callbacks = [];
    const resolved_promise = Promise.resolve();
    let update_scheduled = false;
    function schedule_update() {
        if (!update_scheduled) {
            update_scheduled = true;
            resolved_promise.then(flush);
        }
    }
    function add_render_callback(fn) {
        render_callbacks.push(fn);
    }
    function add_flush_callback(fn) {
        flush_callbacks.push(fn);
    }
    // flush() calls callbacks in this order:
    // 1. All beforeUpdate callbacks, in order: parents before children
    // 2. All bind:this callbacks, in reverse order: children before parents.
    // 3. All afterUpdate callbacks, in order: parents before children. EXCEPT
    //    for afterUpdates called during the initial onMount, which are called in
    //    reverse order: children before parents.
    // Since callbacks might update component values, which could trigger another
    // call to flush(), the following steps guard against this:
    // 1. During beforeUpdate, any updated components will be added to the
    //    dirty_components array and will cause a reentrant call to flush(). Because
    //    the flush index is kept outside the function, the reentrant call will pick
    //    up where the earlier call left off and go through all dirty components. The
    //    current_component value is saved and restored so that the reentrant call will
    //    not interfere with the "parent" flush() call.
    // 2. bind:this callbacks cannot trigger new flush() calls.
    // 3. During afterUpdate, any updated components will NOT have their afterUpdate
    //    callback called a second time; the seen_callbacks set, outside the flush()
    //    function, guarantees this behavior.
    const seen_callbacks = new Set();
    let flushidx = 0; // Do *not* move this inside the flush() function
    function flush() {
        const saved_component = current_component;
        do {
            // first, call beforeUpdate functions
            // and update components
            while (flushidx < dirty_components.length) {
                const component = dirty_components[flushidx];
                flushidx++;
                set_current_component(component);
                update(component.$$);
            }
            set_current_component(null);
            dirty_components.length = 0;
            flushidx = 0;
            while (binding_callbacks.length)
                binding_callbacks.pop()();
            // then, once components are updated, call
            // afterUpdate functions. This may cause
            // subsequent updates...
            for (let i = 0; i < render_callbacks.length; i += 1) {
                const callback = render_callbacks[i];
                if (!seen_callbacks.has(callback)) {
                    // ...so guard against infinite loops
                    seen_callbacks.add(callback);
                    callback();
                }
            }
            render_callbacks.length = 0;
        } while (dirty_components.length);
        while (flush_callbacks.length) {
            flush_callbacks.pop()();
        }
        update_scheduled = false;
        seen_callbacks.clear();
        set_current_component(saved_component);
    }
    function update($$) {
        if ($$.fragment !== null) {
            $$.update();
            run_all($$.before_update);
            const dirty = $$.dirty;
            $$.dirty = [-1];
            $$.fragment && $$.fragment.p($$.ctx, dirty);
            $$.after_update.forEach(add_render_callback);
        }
    }
    const outroing = new Set();
    let outros;
    function group_outros() {
        outros = {
            r: 0,
            c: [],
            p: outros // parent group
        };
    }
    function check_outros() {
        if (!outros.r) {
            run_all(outros.c);
        }
        outros = outros.p;
    }
    function transition_in(block, local) {
        if (block && block.i) {
            outroing.delete(block);
            block.i(local);
        }
    }
    function transition_out(block, local, detach, callback) {
        if (block && block.o) {
            if (outroing.has(block))
                return;
            outroing.add(block);
            outros.c.push(() => {
                outroing.delete(block);
                if (callback) {
                    if (detach)
                        block.d(1);
                    callback();
                }
            });
            block.o(local);
        }
        else if (callback) {
            callback();
        }
    }

    const globals = (typeof window !== 'undefined'
        ? window
        : typeof globalThis !== 'undefined'
            ? globalThis
            : global);

    function get_spread_update(levels, updates) {
        const update = {};
        const to_null_out = {};
        const accounted_for = { $$scope: 1 };
        let i = levels.length;
        while (i--) {
            const o = levels[i];
            const n = updates[i];
            if (n) {
                for (const key in o) {
                    if (!(key in n))
                        to_null_out[key] = 1;
                }
                for (const key in n) {
                    if (!accounted_for[key]) {
                        update[key] = n[key];
                        accounted_for[key] = 1;
                    }
                }
                levels[i] = n;
            }
            else {
                for (const key in o) {
                    accounted_for[key] = 1;
                }
            }
        }
        for (const key in to_null_out) {
            if (!(key in update))
                update[key] = undefined;
        }
        return update;
    }
    function get_spread_object(spread_props) {
        return typeof spread_props === 'object' && spread_props !== null ? spread_props : {};
    }

    function bind(component, name, callback) {
        const index = component.$$.props[name];
        if (index !== undefined) {
            component.$$.bound[index] = callback;
            callback(component.$$.ctx[index]);
        }
    }
    function create_component(block) {
        block && block.c();
    }
    function claim_component(block, parent_nodes) {
        block && block.l(parent_nodes);
    }
    function mount_component(component, target, anchor, customElement) {
        const { fragment, after_update } = component.$$;
        fragment && fragment.m(target, anchor);
        if (!customElement) {
            // onMount happens before the initial afterUpdate
            add_render_callback(() => {
                const new_on_destroy = component.$$.on_mount.map(run).filter(is_function);
                // if the component was destroyed immediately
                // it will update the `$$.on_destroy` reference to `null`.
                // the destructured on_destroy may still reference to the old array
                if (component.$$.on_destroy) {
                    component.$$.on_destroy.push(...new_on_destroy);
                }
                else {
                    // Edge case - component was destroyed immediately,
                    // most likely as a result of a binding initialising
                    run_all(new_on_destroy);
                }
                component.$$.on_mount = [];
            });
        }
        after_update.forEach(add_render_callback);
    }
    function destroy_component(component, detaching) {
        const $$ = component.$$;
        if ($$.fragment !== null) {
            run_all($$.on_destroy);
            $$.fragment && $$.fragment.d(detaching);
            // TODO null out other refs, including component.$$ (but need to
            // preserve final state?)
            $$.on_destroy = $$.fragment = null;
            $$.ctx = [];
        }
    }
    function make_dirty(component, i) {
        if (component.$$.dirty[0] === -1) {
            dirty_components.push(component);
            schedule_update();
            component.$$.dirty.fill(0);
        }
        component.$$.dirty[(i / 31) | 0] |= (1 << (i % 31));
    }
    function init(component, options, instance, create_fragment, not_equal, props, append_styles, dirty = [-1]) {
        const parent_component = current_component;
        set_current_component(component);
        const $$ = component.$$ = {
            fragment: null,
            ctx: [],
            // state
            props,
            update: noop,
            not_equal,
            bound: blank_object(),
            // lifecycle
            on_mount: [],
            on_destroy: [],
            on_disconnect: [],
            before_update: [],
            after_update: [],
            context: new Map(options.context || (parent_component ? parent_component.$$.context : [])),
            // everything else
            callbacks: blank_object(),
            dirty,
            skip_bound: false,
            root: options.target || parent_component.$$.root
        };
        append_styles && append_styles($$.root);
        let ready = false;
        $$.ctx = instance
            ? instance(component, options.props || {}, (i, ret, ...rest) => {
                const value = rest.length ? rest[0] : ret;
                if ($$.ctx && not_equal($$.ctx[i], $$.ctx[i] = value)) {
                    if (!$$.skip_bound && $$.bound[i])
                        $$.bound[i](value);
                    if (ready)
                        make_dirty(component, i);
                }
                return ret;
            })
            : [];
        $$.update();
        ready = true;
        run_all($$.before_update);
        // `false` as a special case of no DOM component
        $$.fragment = create_fragment ? create_fragment($$.ctx) : false;
        if (options.target) {
            if (options.hydrate) {
                start_hydrating();
                const nodes = children(options.target);
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                $$.fragment && $$.fragment.l(nodes);
                nodes.forEach(detach);
            }
            else {
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                $$.fragment && $$.fragment.c();
            }
            if (options.intro)
                transition_in(component.$$.fragment);
            mount_component(component, options.target, options.anchor, options.customElement);
            end_hydrating();
            flush();
        }
        set_current_component(parent_component);
    }
    /**
     * Base class for Svelte components. Used when dev=false.
     */
    class SvelteComponent {
        $destroy() {
            destroy_component(this, 1);
            this.$destroy = noop;
        }
        $on(type, callback) {
            if (!is_function(callback)) {
                return noop;
            }
            const callbacks = (this.$$.callbacks[type] || (this.$$.callbacks[type] = []));
            callbacks.push(callback);
            return () => {
                const index = callbacks.indexOf(callback);
                if (index !== -1)
                    callbacks.splice(index, 1);
            };
        }
        $set($$props) {
            if (this.$$set && !is_empty($$props)) {
                this.$$.skip_bound = true;
                this.$$set($$props);
                this.$$.skip_bound = false;
            }
        }
    }

    function dispatch_dev(type, detail) {
        document.dispatchEvent(custom_event(type, Object.assign({ version: '3.53.1' }, detail), { bubbles: true }));
    }
    function append_hydration_dev(target, node) {
        dispatch_dev('SvelteDOMInsert', { target, node });
        append_hydration(target, node);
    }
    function insert_hydration_dev(target, node, anchor) {
        dispatch_dev('SvelteDOMInsert', { target, node, anchor });
        insert_hydration(target, node, anchor);
    }
    function detach_dev(node) {
        dispatch_dev('SvelteDOMRemove', { node });
        detach(node);
    }
    function listen_dev(node, event, handler, options, has_prevent_default, has_stop_propagation) {
        const modifiers = options === true ? ['capture'] : options ? Array.from(Object.keys(options)) : [];
        if (has_prevent_default)
            modifiers.push('preventDefault');
        if (has_stop_propagation)
            modifiers.push('stopPropagation');
        dispatch_dev('SvelteDOMAddEventListener', { node, event, handler, modifiers });
        const dispose = listen(node, event, handler, options);
        return () => {
            dispatch_dev('SvelteDOMRemoveEventListener', { node, event, handler, modifiers });
            dispose();
        };
    }
    function attr_dev(node, attribute, value) {
        attr(node, attribute, value);
        if (value == null)
            dispatch_dev('SvelteDOMRemoveAttribute', { node, attribute });
        else
            dispatch_dev('SvelteDOMSetAttribute', { node, attribute, value });
    }
    function set_data_dev(text, data) {
        data = '' + data;
        if (text.wholeText === data)
            return;
        dispatch_dev('SvelteDOMSetData', { node: text, data });
        text.data = data;
    }
    function validate_each_argument(arg) {
        if (typeof arg !== 'string' && !(arg && typeof arg === 'object' && 'length' in arg)) {
            let msg = '{#each} only iterates over array-like objects.';
            if (typeof Symbol === 'function' && arg && Symbol.iterator in arg) {
                msg += ' You can use a spread to convert this iterable into an array.';
            }
            throw new Error(msg);
        }
    }
    function validate_slots(name, slot, keys) {
        for (const slot_key of Object.keys(slot)) {
            if (!~keys.indexOf(slot_key)) {
                console.warn(`<${name}> received an unexpected slot "${slot_key}".`);
            }
        }
    }
    function construct_svelte_component_dev(component, props) {
        const error_message = 'this={...} of <svelte:component> should specify a Svelte component.';
        try {
            const instance = new component(props);
            if (!instance.$$ || !instance.$set || !instance.$on || !instance.$destroy) {
                throw new Error(error_message);
            }
            return instance;
        }
        catch (err) {
            const { message } = err;
            if (typeof message === 'string' && message.indexOf('is not a constructor') !== -1) {
                throw new Error(error_message);
            }
            else {
                throw err;
            }
        }
    }
    /**
     * Base class for Svelte components with some minor dev-enhancements. Used when dev=true.
     */
    class SvelteComponentDev extends SvelteComponent {
        constructor(options) {
            if (!options || (!options.target && !options.$$inline)) {
                throw new Error("'target' is a required option");
            }
            super();
        }
        $destroy() {
            super.$destroy();
            this.$destroy = () => {
                console.warn('Component was already destroyed'); // eslint-disable-line no-console
            };
        }
        $capture_state() { }
        $inject_state() { }
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

    function create_fragment$t(ctx) {
    	let current;
    	const default_slot_template = /*#slots*/ ctx[9].default;
    	const default_slot = create_slot(default_slot_template, ctx, /*$$scope*/ ctx[8], null);

    	const block = {
    		c: function create() {
    			if (default_slot) default_slot.c();
    		},
    		l: function claim(nodes) {
    			if (default_slot) default_slot.l(nodes);
    		},
    		m: function mount(target, anchor) {
    			if (default_slot) {
    				default_slot.m(target, anchor);
    			}

    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			if (default_slot) {
    				if (default_slot.p && (!current || dirty & /*$$scope*/ 256)) {
    					update_slot_base(
    						default_slot,
    						default_slot_template,
    						ctx,
    						/*$$scope*/ ctx[8],
    						!current
    						? get_all_dirty_from_scope(/*$$scope*/ ctx[8])
    						: get_slot_changes(default_slot_template, /*$$scope*/ ctx[8], dirty, null),
    						null
    					);
    				}
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(default_slot, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(default_slot, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (default_slot) default_slot.d(detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$t.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$t($$self, $$props, $$invalidate) {
    	let $location;
    	let $routes;
    	let $base;
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Router', slots, ['default']);
    	let { basepath = "/" } = $$props;
    	let { url = null } = $$props;
    	const locationContext = getContext(LOCATION);
    	const routerContext = getContext(ROUTER);
    	const routes = writable([]);
    	validate_store(routes, 'routes');
    	component_subscribe($$self, routes, value => $$invalidate(6, $routes = value));
    	const activeRoute = writable(null);
    	let hasActiveRoute = false; // Used in SSR to synchronously set that a Route is active.

    	// If locationContext is not set, this is the topmost Router in the tree.
    	// If the `url` prop is given we force the location to it.
    	const location = locationContext || writable(url ? { pathname: url } : globalHistory.location);

    	validate_store(location, 'location');
    	component_subscribe($$self, location, value => $$invalidate(5, $location = value));

    	// If routerContext is set, the routerBase of the parent Router
    	// will be the base for this Router's descendants.
    	// If routerContext is not set, the path and resolved uri will both
    	// have the value of the basepath prop.
    	const base = routerContext
    	? routerContext.routerBase
    	: writable({ path: basepath, uri: basepath });

    	validate_store(base, 'base');
    	component_subscribe($$self, base, value => $$invalidate(7, $base = value));

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

    	const writable_props = ['basepath', 'url'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Router> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('basepath' in $$props) $$invalidate(3, basepath = $$props.basepath);
    		if ('url' in $$props) $$invalidate(4, url = $$props.url);
    		if ('$$scope' in $$props) $$invalidate(8, $$scope = $$props.$$scope);
    	};

    	$$self.$capture_state = () => ({
    		getContext,
    		setContext,
    		onMount,
    		writable,
    		derived,
    		LOCATION,
    		ROUTER,
    		globalHistory,
    		pick,
    		match,
    		stripSlashes,
    		combinePaths,
    		basepath,
    		url,
    		locationContext,
    		routerContext,
    		routes,
    		activeRoute,
    		hasActiveRoute,
    		location,
    		base,
    		routerBase,
    		registerRoute,
    		unregisterRoute,
    		$location,
    		$routes,
    		$base
    	});

    	$$self.$inject_state = $$props => {
    		if ('basepath' in $$props) $$invalidate(3, basepath = $$props.basepath);
    		if ('url' in $$props) $$invalidate(4, url = $$props.url);
    		if ('hasActiveRoute' in $$props) hasActiveRoute = $$props.hasActiveRoute;
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*$base*/ 128) {
    			// This reactive statement will update all the Routes' path when
    			// the basepath changes.
    			{
    				const { path: basepath } = $base;

    				routes.update(rs => {
    					rs.forEach(r => r.path = combinePaths(basepath, r._path));
    					return rs;
    				});
    			}
    		}

    		if ($$self.$$.dirty & /*$routes, $location*/ 96) {
    			// This reactive statement will be run when the Router is created
    			// when there are no Routes and then again the following tick, so it
    			// will not find an active Route in SSR and in the browser it will only
    			// pick an active Route after all Routes have been registered.
    			{
    				const bestMatch = pick($routes, $location.pathname);
    				activeRoute.set(bestMatch);
    			}
    		}
    	};

    	return [
    		routes,
    		location,
    		base,
    		basepath,
    		url,
    		$location,
    		$routes,
    		$base,
    		$$scope,
    		slots
    	];
    }

    class Router extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$t, create_fragment$t, safe_not_equal, { basepath: 3, url: 4 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Router",
    			options,
    			id: create_fragment$t.name
    		});
    	}

    	get basepath() {
    		throw new Error("<Router>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set basepath(value) {
    		throw new Error("<Router>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get url() {
    		throw new Error("<Router>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set url(value) {
    		throw new Error("<Router>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* node_modules\svelte-routing\src\Route.svelte generated by Svelte v3.53.1 */

    const get_default_slot_changes = dirty => ({
    	params: dirty & /*routeParams*/ 4,
    	location: dirty & /*$location*/ 16
    });

    const get_default_slot_context = ctx => ({
    	params: /*routeParams*/ ctx[2],
    	location: /*$location*/ ctx[4]
    });

    // (40:0) {#if $activeRoute !== null && $activeRoute.route === route}
    function create_if_block$a(ctx) {
    	let current_block_type_index;
    	let if_block;
    	let if_block_anchor;
    	let current;
    	const if_block_creators = [create_if_block_1$2, create_else_block$3];
    	const if_blocks = [];

    	function select_block_type(ctx, dirty) {
    		if (/*component*/ ctx[0] !== null) return 0;
    		return 1;
    	}

    	current_block_type_index = select_block_type(ctx);
    	if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);

    	const block = {
    		c: function create() {
    			if_block.c();
    			if_block_anchor = empty();
    		},
    		l: function claim(nodes) {
    			if_block.l(nodes);
    			if_block_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if_blocks[current_block_type_index].m(target, anchor);
    			insert_hydration_dev(target, if_block_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			let previous_block_index = current_block_type_index;
    			current_block_type_index = select_block_type(ctx);

    			if (current_block_type_index === previous_block_index) {
    				if_blocks[current_block_type_index].p(ctx, dirty);
    			} else {
    				group_outros();

    				transition_out(if_blocks[previous_block_index], 1, 1, () => {
    					if_blocks[previous_block_index] = null;
    				});

    				check_outros();
    				if_block = if_blocks[current_block_type_index];

    				if (!if_block) {
    					if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);
    					if_block.c();
    				} else {
    					if_block.p(ctx, dirty);
    				}

    				transition_in(if_block, 1);
    				if_block.m(if_block_anchor.parentNode, if_block_anchor);
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(if_block);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if_blocks[current_block_type_index].d(detaching);
    			if (detaching) detach_dev(if_block_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$a.name,
    		type: "if",
    		source: "(40:0) {#if $activeRoute !== null && $activeRoute.route === route}",
    		ctx
    	});

    	return block;
    }

    // (43:2) {:else}
    function create_else_block$3(ctx) {
    	let current;
    	const default_slot_template = /*#slots*/ ctx[10].default;
    	const default_slot = create_slot(default_slot_template, ctx, /*$$scope*/ ctx[9], get_default_slot_context);

    	const block = {
    		c: function create() {
    			if (default_slot) default_slot.c();
    		},
    		l: function claim(nodes) {
    			if (default_slot) default_slot.l(nodes);
    		},
    		m: function mount(target, anchor) {
    			if (default_slot) {
    				default_slot.m(target, anchor);
    			}

    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			if (default_slot) {
    				if (default_slot.p && (!current || dirty & /*$$scope, routeParams, $location*/ 532)) {
    					update_slot_base(
    						default_slot,
    						default_slot_template,
    						ctx,
    						/*$$scope*/ ctx[9],
    						!current
    						? get_all_dirty_from_scope(/*$$scope*/ ctx[9])
    						: get_slot_changes(default_slot_template, /*$$scope*/ ctx[9], dirty, get_default_slot_changes),
    						get_default_slot_context
    					);
    				}
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(default_slot, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(default_slot, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (default_slot) default_slot.d(detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_else_block$3.name,
    		type: "else",
    		source: "(43:2) {:else}",
    		ctx
    	});

    	return block;
    }

    // (41:2) {#if component !== null}
    function create_if_block_1$2(ctx) {
    	let switch_instance;
    	let switch_instance_anchor;
    	let current;

    	const switch_instance_spread_levels = [
    		{ location: /*$location*/ ctx[4] },
    		/*routeParams*/ ctx[2],
    		/*routeProps*/ ctx[3]
    	];

    	var switch_value = /*component*/ ctx[0];

    	function switch_props(ctx) {
    		let switch_instance_props = {};

    		for (let i = 0; i < switch_instance_spread_levels.length; i += 1) {
    			switch_instance_props = assign(switch_instance_props, switch_instance_spread_levels[i]);
    		}

    		return {
    			props: switch_instance_props,
    			$$inline: true
    		};
    	}

    	if (switch_value) {
    		switch_instance = construct_svelte_component_dev(switch_value, switch_props());
    	}

    	const block = {
    		c: function create() {
    			if (switch_instance) create_component(switch_instance.$$.fragment);
    			switch_instance_anchor = empty();
    		},
    		l: function claim(nodes) {
    			if (switch_instance) claim_component(switch_instance.$$.fragment, nodes);
    			switch_instance_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if (switch_instance) mount_component(switch_instance, target, anchor);
    			insert_hydration_dev(target, switch_instance_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			const switch_instance_changes = (dirty & /*$location, routeParams, routeProps*/ 28)
    			? get_spread_update(switch_instance_spread_levels, [
    					dirty & /*$location*/ 16 && { location: /*$location*/ ctx[4] },
    					dirty & /*routeParams*/ 4 && get_spread_object(/*routeParams*/ ctx[2]),
    					dirty & /*routeProps*/ 8 && get_spread_object(/*routeProps*/ ctx[3])
    				])
    			: {};

    			if (switch_value !== (switch_value = /*component*/ ctx[0])) {
    				if (switch_instance) {
    					group_outros();
    					const old_component = switch_instance;

    					transition_out(old_component.$$.fragment, 1, 0, () => {
    						destroy_component(old_component, 1);
    					});

    					check_outros();
    				}

    				if (switch_value) {
    					switch_instance = construct_svelte_component_dev(switch_value, switch_props());
    					create_component(switch_instance.$$.fragment);
    					transition_in(switch_instance.$$.fragment, 1);
    					mount_component(switch_instance, switch_instance_anchor.parentNode, switch_instance_anchor);
    				} else {
    					switch_instance = null;
    				}
    			} else if (switch_value) {
    				switch_instance.$set(switch_instance_changes);
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			if (switch_instance) transition_in(switch_instance.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			if (switch_instance) transition_out(switch_instance.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(switch_instance_anchor);
    			if (switch_instance) destroy_component(switch_instance, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1$2.name,
    		type: "if",
    		source: "(41:2) {#if component !== null}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$s(ctx) {
    	let if_block_anchor;
    	let current;
    	let if_block = /*$activeRoute*/ ctx[1] !== null && /*$activeRoute*/ ctx[1].route === /*route*/ ctx[7] && create_if_block$a(ctx);

    	const block = {
    		c: function create() {
    			if (if_block) if_block.c();
    			if_block_anchor = empty();
    		},
    		l: function claim(nodes) {
    			if (if_block) if_block.l(nodes);
    			if_block_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if (if_block) if_block.m(target, anchor);
    			insert_hydration_dev(target, if_block_anchor, anchor);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*$activeRoute*/ ctx[1] !== null && /*$activeRoute*/ ctx[1].route === /*route*/ ctx[7]) {
    				if (if_block) {
    					if_block.p(ctx, dirty);

    					if (dirty & /*$activeRoute*/ 2) {
    						transition_in(if_block, 1);
    					}
    				} else {
    					if_block = create_if_block$a(ctx);
    					if_block.c();
    					transition_in(if_block, 1);
    					if_block.m(if_block_anchor.parentNode, if_block_anchor);
    				}
    			} else if (if_block) {
    				group_outros();

    				transition_out(if_block, 1, 1, () => {
    					if_block = null;
    				});

    				check_outros();
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(if_block);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (if_block) if_block.d(detaching);
    			if (detaching) detach_dev(if_block_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$s.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$s($$self, $$props, $$invalidate) {
    	let $activeRoute;
    	let $location;
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Route', slots, ['default']);
    	let { path = "" } = $$props;
    	let { component = null } = $$props;
    	const { registerRoute, unregisterRoute, activeRoute } = getContext(ROUTER);
    	validate_store(activeRoute, 'activeRoute');
    	component_subscribe($$self, activeRoute, value => $$invalidate(1, $activeRoute = value));
    	const location = getContext(LOCATION);
    	validate_store(location, 'location');
    	component_subscribe($$self, location, value => $$invalidate(4, $location = value));

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

    	$$self.$$set = $$new_props => {
    		$$invalidate(13, $$props = assign(assign({}, $$props), exclude_internal_props($$new_props)));
    		if ('path' in $$new_props) $$invalidate(8, path = $$new_props.path);
    		if ('component' in $$new_props) $$invalidate(0, component = $$new_props.component);
    		if ('$$scope' in $$new_props) $$invalidate(9, $$scope = $$new_props.$$scope);
    	};

    	$$self.$capture_state = () => ({
    		getContext,
    		onDestroy,
    		ROUTER,
    		LOCATION,
    		path,
    		component,
    		registerRoute,
    		unregisterRoute,
    		activeRoute,
    		location,
    		route,
    		routeParams,
    		routeProps,
    		$activeRoute,
    		$location
    	});

    	$$self.$inject_state = $$new_props => {
    		$$invalidate(13, $$props = assign(assign({}, $$props), $$new_props));
    		if ('path' in $$props) $$invalidate(8, path = $$new_props.path);
    		if ('component' in $$props) $$invalidate(0, component = $$new_props.component);
    		if ('routeParams' in $$props) $$invalidate(2, routeParams = $$new_props.routeParams);
    		if ('routeProps' in $$props) $$invalidate(3, routeProps = $$new_props.routeProps);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*$activeRoute*/ 2) {
    			if ($activeRoute && $activeRoute.route === route) {
    				$$invalidate(2, routeParams = $activeRoute.params);
    			}
    		}

    		{
    			const { path, component, ...rest } = $$props;
    			$$invalidate(3, routeProps = rest);
    		}
    	};

    	$$props = exclude_internal_props($$props);

    	return [
    		component,
    		$activeRoute,
    		routeParams,
    		routeProps,
    		$location,
    		activeRoute,
    		location,
    		route,
    		path,
    		$$scope,
    		slots
    	];
    }

    class Route extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$s, create_fragment$s, safe_not_equal, { path: 8, component: 0 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Route",
    			options,
    			id: create_fragment$s.name
    		});
    	}

    	get path() {
    		throw new Error("<Route>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set path(value) {
    		throw new Error("<Route>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get component() {
    		throw new Error("<Route>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set component(value) {
    		throw new Error("<Route>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\Footer.svelte generated by Svelte v3.53.1 */

    const file$q = "src\\components\\Footer.svelte";

    function create_fragment$r(ctx) {
    	let div1;
    	let a;
    	let img0;
    	let img0_src_value;
    	let t0;
    	let div0;
    	let t1;
    	let t2;
    	let img1;
    	let img1_src_value;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			a = element("a");
    			img0 = element("img");
    			t0 = space();
    			div0 = element("div");
    			t1 = text("API Analytics");
    			t2 = space();
    			img1 = element("img");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);

    			a = claim_element(div1_nodes, "A", {
    				class: true,
    				rel: true,
    				target: true,
    				href: true
    			});

    			var a_nodes = children(a);

    			img0 = claim_element(a_nodes, "IMG", {
    				class: true,
    				height: true,
    				src: true,
    				alt: true
    			});

    			a_nodes.forEach(detach_dev);
    			t0 = claim_space(div1_nodes);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t1 = claim_text(div0_nodes, "API Analytics");
    			div0_nodes.forEach(detach_dev);
    			t2 = claim_space(div1_nodes);
    			img1 = claim_element(div1_nodes, "IMG", { class: true, src: true, alt: true });
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(img0, "class", "github-logo svelte-82t1ww");
    			attr_dev(img0, "height", "30px");
    			if (!src_url_equal(img0.src, img0_src_value = "../img/github.png")) attr_dev(img0, "src", img0_src_value);
    			attr_dev(img0, "alt", "");
    			add_location(img0, file$q, 2, 4, 137);
    			attr_dev(a, "class", "github-link");
    			attr_dev(a, "rel", "noreferrer");
    			attr_dev(a, "target", "_blank");
    			attr_dev(a, "href", "https://github.com/tom-draper/api-analytics");
    			add_location(a, file$q, 1, 2, 24);
    			attr_dev(div0, "class", "logo svelte-82t1ww");
    			add_location(div0, file$q, 4, 2, 221);
    			attr_dev(img1, "class", "footer-logo");
    			if (!src_url_equal(img1.src, img1_src_value = "../img/logo.png")) attr_dev(img1, "src", img1_src_value);
    			attr_dev(img1, "alt", "");
    			add_location(img1, file$q, 5, 2, 262);
    			attr_dev(div1, "class", "footer svelte-82t1ww");
    			add_location(div1, file$q, 0, 0, 0);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, a);
    			append_hydration_dev(a, img0);
    			append_hydration_dev(div1, t0);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t1);
    			append_hydration_dev(div1, t2);
    			append_hydration_dev(div1, img1);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$r.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$r($$self, $$props) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Footer', slots, []);
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Footer> was created with unknown prop '${key}'`);
    	});

    	return [];
    }

    class Footer extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$r, create_fragment$r, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Footer",
    			options,
    			id: create_fragment$r.name
    		});
    	}
    }

    /* src\routes\Home.svelte generated by Svelte v3.53.1 */
    const file$p = "src\\routes\\Home.svelte";

    function create_fragment$q(ctx) {
    	let div19;
    	let div6;
    	let div5;
    	let div3;
    	let h10;
    	let t0;
    	let t1;
    	let h2;
    	let t2;
    	let t3;
    	let div2;
    	let a0;
    	let div0;
    	let t4;
    	let span;
    	let t5;
    	let t6;
    	let a1;
    	let div1;
    	let t7;
    	let t8;
    	let div4;
    	let img0;
    	let img0_src_value;
    	let t9;
    	let div12;
    	let div7;
    	let img1;
    	let img1_src_value;
    	let t10;
    	let h11;
    	let t11;
    	let t12;
    	let div11;
    	let div8;
    	let t13;
    	let t14;
    	let div10;
    	let a2;
    	let div9;
    	let t15;
    	let t16;
    	let img2;
    	let img2_src_value;
    	let t17;
    	let div18;
    	let div13;
    	let img3;
    	let img3_src_value;
    	let t18;
    	let h12;
    	let t19;
    	let t20;
    	let div17;
    	let div14;
    	let t21;
    	let t22;
    	let div16;
    	let a3;
    	let div15;
    	let t23;
    	let t24;
    	let img4;
    	let img4_src_value;
    	let t25;
    	let footer;
    	let current;
    	footer = new Footer({ $$inline: true });

    	const block = {
    		c: function create() {
    			div19 = element("div");
    			div6 = element("div");
    			div5 = element("div");
    			div3 = element("div");
    			h10 = element("h1");
    			t0 = text("API Analytics");
    			t1 = space();
    			h2 = element("h2");
    			t2 = text("Monitoring and analytics for API frameworks.");
    			t3 = space();
    			div2 = element("div");
    			a0 = element("a");
    			div0 = element("div");
    			t4 = text("Try now - it's ");
    			span = element("span");
    			t5 = text("free");
    			t6 = space();
    			a1 = element("a");
    			div1 = element("div");
    			t7 = text("Demo");
    			t8 = space();
    			div4 = element("div");
    			img0 = element("img");
    			t9 = space();
    			div12 = element("div");
    			div7 = element("div");
    			img1 = element("img");
    			t10 = space();
    			h11 = element("h1");
    			t11 = text("Dashboard");
    			t12 = space();
    			div11 = element("div");
    			div8 = element("div");
    			t13 = text("An all-in-one analytics dashboard. Real-time insight into your API's\r\n        usage.");
    			t14 = space();
    			div10 = element("div");
    			a2 = element("a");
    			div9 = element("div");
    			t15 = text("Open");
    			t16 = space();
    			img2 = element("img");
    			t17 = space();
    			div18 = element("div");
    			div13 = element("div");
    			img3 = element("img");
    			t18 = space();
    			h12 = element("h1");
    			t19 = text("Monitoring");
    			t20 = space();
    			div17 = element("div");
    			div14 = element("div");
    			t21 = text("Active monitoring and error notifications. Peace of mind.");
    			t22 = space();
    			div16 = element("div");
    			a3 = element("a");
    			div15 = element("div");
    			t23 = text("Coming Soon");
    			t24 = space();
    			img4 = element("img");
    			t25 = space();
    			create_component(footer.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div19 = claim_element(nodes, "DIV", { class: true });
    			var div19_nodes = children(div19);
    			div6 = claim_element(div19_nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			div3 = claim_element(div5_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			h10 = claim_element(div3_nodes, "H1", { class: true });
    			var h10_nodes = children(h10);
    			t0 = claim_text(h10_nodes, "API Analytics");
    			h10_nodes.forEach(detach_dev);
    			t1 = claim_space(div3_nodes);
    			h2 = claim_element(div3_nodes, "H2", { class: true });
    			var h2_nodes = children(h2);
    			t2 = claim_text(h2_nodes, "Monitoring and analytics for API frameworks.");
    			h2_nodes.forEach(detach_dev);
    			t3 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			a0 = claim_element(div2_nodes, "A", { href: true, class: true });
    			var a0_nodes = children(a0);
    			div0 = claim_element(a0_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t4 = claim_text(div0_nodes, "Try now - it's ");
    			span = claim_element(div0_nodes, "SPAN", { class: true });
    			var span_nodes = children(span);
    			t5 = claim_text(span_nodes, "free");
    			span_nodes.forEach(detach_dev);
    			div0_nodes.forEach(detach_dev);
    			a0_nodes.forEach(detach_dev);
    			t6 = claim_space(div2_nodes);
    			a1 = claim_element(div2_nodes, "A", { href: true, class: true });
    			var a1_nodes = children(a1);
    			div1 = claim_element(a1_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			t7 = claim_text(div1_nodes, "Demo");
    			div1_nodes.forEach(detach_dev);
    			a1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			t8 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			img0 = claim_element(div4_nodes, "IMG", { class: true, src: true, alt: true });
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			t9 = claim_space(div19_nodes);
    			div12 = claim_element(div19_nodes, "DIV", { class: true });
    			var div12_nodes = children(div12);
    			div7 = claim_element(div12_nodes, "DIV", { class: true });
    			var div7_nodes = children(div7);
    			img1 = claim_element(div7_nodes, "IMG", { class: true, src: true, alt: true });
    			t10 = claim_space(div7_nodes);
    			h11 = claim_element(div7_nodes, "H1", { class: true });
    			var h11_nodes = children(h11);
    			t11 = claim_text(h11_nodes, "Dashboard");
    			h11_nodes.forEach(detach_dev);
    			div7_nodes.forEach(detach_dev);
    			t12 = claim_space(div12_nodes);
    			div11 = claim_element(div12_nodes, "DIV", { class: true });
    			var div11_nodes = children(div11);
    			div8 = claim_element(div11_nodes, "DIV", { class: true });
    			var div8_nodes = children(div8);
    			t13 = claim_text(div8_nodes, "An all-in-one analytics dashboard. Real-time insight into your API's\r\n        usage.");
    			div8_nodes.forEach(detach_dev);
    			t14 = claim_space(div11_nodes);
    			div10 = claim_element(div11_nodes, "DIV", { class: true });
    			var div10_nodes = children(div10);
    			a2 = claim_element(div10_nodes, "A", { href: true, class: true });
    			var a2_nodes = children(a2);
    			div9 = claim_element(a2_nodes, "DIV", { class: true });
    			var div9_nodes = children(div9);
    			t15 = claim_text(div9_nodes, "Open");
    			div9_nodes.forEach(detach_dev);
    			a2_nodes.forEach(detach_dev);
    			div10_nodes.forEach(detach_dev);
    			div11_nodes.forEach(detach_dev);
    			t16 = claim_space(div12_nodes);
    			img2 = claim_element(div12_nodes, "IMG", { class: true, src: true, alt: true });
    			div12_nodes.forEach(detach_dev);
    			t17 = claim_space(div19_nodes);
    			div18 = claim_element(div19_nodes, "DIV", { class: true });
    			var div18_nodes = children(div18);
    			div13 = claim_element(div18_nodes, "DIV", { class: true });
    			var div13_nodes = children(div13);
    			img3 = claim_element(div13_nodes, "IMG", { class: true, src: true, alt: true });
    			t18 = claim_space(div13_nodes);
    			h12 = claim_element(div13_nodes, "H1", { class: true });
    			var h12_nodes = children(h12);
    			t19 = claim_text(h12_nodes, "Monitoring");
    			h12_nodes.forEach(detach_dev);
    			div13_nodes.forEach(detach_dev);
    			t20 = claim_space(div18_nodes);
    			div17 = claim_element(div18_nodes, "DIV", { class: true });
    			var div17_nodes = children(div17);
    			div14 = claim_element(div17_nodes, "DIV", { class: true });
    			var div14_nodes = children(div14);
    			t21 = claim_text(div14_nodes, "Active monitoring and error notifications. Peace of mind.");
    			div14_nodes.forEach(detach_dev);
    			t22 = claim_space(div17_nodes);
    			div16 = claim_element(div17_nodes, "DIV", { class: true });
    			var div16_nodes = children(div16);
    			a3 = claim_element(div16_nodes, "A", { href: true, class: true });
    			var a3_nodes = children(a3);
    			div15 = claim_element(a3_nodes, "DIV", { class: true });
    			var div15_nodes = children(div15);
    			t23 = claim_text(div15_nodes, "Coming Soon");
    			div15_nodes.forEach(detach_dev);
    			a3_nodes.forEach(detach_dev);
    			div16_nodes.forEach(detach_dev);
    			div17_nodes.forEach(detach_dev);
    			t24 = claim_space(div18_nodes);
    			img4 = claim_element(div18_nodes, "IMG", { class: true, src: true, alt: true });
    			div18_nodes.forEach(detach_dev);
    			div19_nodes.forEach(detach_dev);
    			t25 = claim_space(nodes);
    			claim_component(footer.$$.fragment, nodes);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(h10, "class", "svelte-1m4i1ue");
    			add_location(h10, file$p, 9, 8, 230);
    			attr_dev(h2, "class", "svelte-1m4i1ue");
    			add_location(h2, file$p, 10, 8, 262);
    			attr_dev(span, "class", "italic svelte-1m4i1ue");
    			add_location(span, file$p, 14, 29, 452);
    			attr_dev(div0, "class", "text");
    			add_location(div0, file$p, 13, 12, 403);
    			attr_dev(a0, "href", "/generate");
    			attr_dev(a0, "class", "link svelte-1m4i1ue");
    			add_location(a0, file$p, 12, 10, 356);
    			attr_dev(div1, "class", "text");
    			add_location(div1, file$p, 18, 12, 595);
    			attr_dev(a1, "href", "/dashboard/demo");
    			attr_dev(a1, "class", "link secondary svelte-1m4i1ue");
    			add_location(a1, file$p, 17, 10, 532);
    			attr_dev(div2, "class", "links svelte-1m4i1ue");
    			add_location(div2, file$p, 11, 8, 325);
    			attr_dev(div3, "class", "left svelte-1m4i1ue");
    			add_location(div3, file$p, 8, 6, 202);
    			attr_dev(img0, "class", "logo svelte-1m4i1ue");
    			if (!src_url_equal(img0.src, img0_src_value = "img/home-logo2.png")) attr_dev(img0, "src", img0_src_value);
    			attr_dev(img0, "alt", "");
    			add_location(img0, file$p, 23, 8, 706);
    			attr_dev(div4, "class", "right svelte-1m4i1ue");
    			add_location(div4, file$p, 22, 6, 677);
    			attr_dev(div5, "class", "landing-page svelte-1m4i1ue");
    			add_location(div5, file$p, 7, 4, 168);
    			attr_dev(div6, "class", "landing-page-container svelte-1m4i1ue");
    			add_location(div6, file$p, 6, 2, 126);
    			attr_dev(img1, "class", "lightning-top svelte-1m4i1ue");
    			if (!src_url_equal(img1.src, img1_src_value = "img/logo.png")) attr_dev(img1, "src", img1_src_value);
    			attr_dev(img1, "alt", "");
    			add_location(img1, file$p, 47, 6, 1493);
    			attr_dev(h11, "class", "dashboard-title svelte-1m4i1ue");
    			add_location(h11, file$p, 48, 6, 1556);
    			attr_dev(div7, "class", "dashboard-title-container svelte-1m4i1ue");
    			add_location(div7, file$p, 46, 4, 1446);
    			attr_dev(div8, "class", "dashboard-content-text svelte-1m4i1ue");
    			add_location(div8, file$p, 51, 6, 1655);
    			attr_dev(div9, "class", "dashboard-btn-text svelte-1m4i1ue");
    			add_location(div9, file$p, 57, 10, 1919);
    			attr_dev(a2, "href", "/dashboard");
    			attr_dev(a2, "class", "dashboard-btn secondary svelte-1m4i1ue");
    			add_location(a2, file$p, 56, 8, 1854);
    			attr_dev(div10, "class", "dashboard-btn-container svelte-1m4i1ue");
    			add_location(div10, file$p, 55, 6, 1807);
    			attr_dev(div11, "class", "dashboard-content svelte-1m4i1ue");
    			add_location(div11, file$p, 50, 4, 1616);
    			attr_dev(img2, "class", "dashboard-img svelte-1m4i1ue");
    			if (!src_url_equal(img2.src, img2_src_value = "img/dashboard.png")) attr_dev(img2, "src", img2_src_value);
    			attr_dev(img2, "alt", "");
    			add_location(img2, file$p, 61, 4, 2007);
    			attr_dev(div12, "class", "dashboard svelte-1m4i1ue");
    			add_location(div12, file$p, 45, 2, 1417);
    			attr_dev(img3, "class", "lightning-top svelte-1m4i1ue");
    			if (!src_url_equal(img3.src, img3_src_value = "img/logo.png")) attr_dev(img3, "src", img3_src_value);
    			attr_dev(img3, "alt", "");
    			add_location(img3, file$p, 65, 6, 2157);
    			attr_dev(h12, "class", "dashboard-title svelte-1m4i1ue");
    			add_location(h12, file$p, 66, 6, 2220);
    			attr_dev(div13, "class", "dashboard-title-container svelte-1m4i1ue");
    			add_location(div13, file$p, 64, 4, 2110);
    			attr_dev(div14, "class", "dashboard-content-text svelte-1m4i1ue");
    			add_location(div14, file$p, 69, 6, 2320);
    			attr_dev(div15, "class", "dashboard-btn-text svelte-1m4i1ue");
    			add_location(div15, file$p, 74, 10, 2548);
    			attr_dev(a3, "href", "/");
    			attr_dev(a3, "class", "dashboard-btn secondary svelte-1m4i1ue");
    			add_location(a3, file$p, 73, 8, 2492);
    			attr_dev(div16, "class", "dashboard-btn-container svelte-1m4i1ue");
    			add_location(div16, file$p, 72, 6, 2445);
    			attr_dev(div17, "class", "dashboard-content svelte-1m4i1ue");
    			add_location(div17, file$p, 68, 4, 2281);
    			attr_dev(img4, "class", "dashboard-img svelte-1m4i1ue");
    			if (!src_url_equal(img4.src, img4_src_value = "img/monitoring.png")) attr_dev(img4, "src", img4_src_value);
    			attr_dev(img4, "alt", "");
    			add_location(img4, file$p, 78, 4, 2643);
    			attr_dev(div18, "class", "dashboard svelte-1m4i1ue");
    			add_location(div18, file$p, 63, 2, 2081);
    			attr_dev(div19, "class", "home svelte-1m4i1ue");
    			add_location(div19, file$p, 5, 0, 104);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div19, anchor);
    			append_hydration_dev(div19, div6);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, div3);
    			append_hydration_dev(div3, h10);
    			append_hydration_dev(h10, t0);
    			append_hydration_dev(div3, t1);
    			append_hydration_dev(div3, h2);
    			append_hydration_dev(h2, t2);
    			append_hydration_dev(div3, t3);
    			append_hydration_dev(div3, div2);
    			append_hydration_dev(div2, a0);
    			append_hydration_dev(a0, div0);
    			append_hydration_dev(div0, t4);
    			append_hydration_dev(div0, span);
    			append_hydration_dev(span, t5);
    			append_hydration_dev(div2, t6);
    			append_hydration_dev(div2, a1);
    			append_hydration_dev(a1, div1);
    			append_hydration_dev(div1, t7);
    			append_hydration_dev(div5, t8);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, img0);
    			append_hydration_dev(div19, t9);
    			append_hydration_dev(div19, div12);
    			append_hydration_dev(div12, div7);
    			append_hydration_dev(div7, img1);
    			append_hydration_dev(div7, t10);
    			append_hydration_dev(div7, h11);
    			append_hydration_dev(h11, t11);
    			append_hydration_dev(div12, t12);
    			append_hydration_dev(div12, div11);
    			append_hydration_dev(div11, div8);
    			append_hydration_dev(div8, t13);
    			append_hydration_dev(div11, t14);
    			append_hydration_dev(div11, div10);
    			append_hydration_dev(div10, a2);
    			append_hydration_dev(a2, div9);
    			append_hydration_dev(div9, t15);
    			append_hydration_dev(div12, t16);
    			append_hydration_dev(div12, img2);
    			append_hydration_dev(div19, t17);
    			append_hydration_dev(div19, div18);
    			append_hydration_dev(div18, div13);
    			append_hydration_dev(div13, img3);
    			append_hydration_dev(div13, t18);
    			append_hydration_dev(div13, h12);
    			append_hydration_dev(h12, t19);
    			append_hydration_dev(div18, t20);
    			append_hydration_dev(div18, div17);
    			append_hydration_dev(div17, div14);
    			append_hydration_dev(div14, t21);
    			append_hydration_dev(div17, t22);
    			append_hydration_dev(div17, div16);
    			append_hydration_dev(div16, a3);
    			append_hydration_dev(a3, div15);
    			append_hydration_dev(div15, t23);
    			append_hydration_dev(div18, t24);
    			append_hydration_dev(div18, img4);
    			insert_hydration_dev(target, t25, anchor);
    			mount_component(footer, target, anchor);
    			current = true;
    		},
    		p: noop,
    		i: function intro(local) {
    			if (current) return;
    			transition_in(footer.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(footer.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div19);
    			if (detaching) detach_dev(t25);
    			destroy_component(footer, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$q.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$q($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Home', slots, []);
    	let framework = 'Django';
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Home> was created with unknown prop '${key}'`);
    	});

    	$$self.$capture_state = () => ({ framework, Footer });

    	$$self.$inject_state = $$props => {
    		if ('framework' in $$props) framework = $$props.framework;
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [];
    }

    class Home extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$q, create_fragment$q, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Home",
    			options,
    			id: create_fragment$q.name
    		});
    	}
    }

    /* src\routes\Generate.svelte generated by Svelte v3.53.1 */

    const file$o = "src\\routes\\Generate.svelte";

    function create_fragment$p(ctx) {
    	let div7;
    	let div3;
    	let h2;
    	let t0;
    	let t1;
    	let input;
    	let t2;
    	let button0;
    	let t3;
    	let t4;
    	let button1;
    	let img0;
    	let img0_src_value;
    	let t5;
    	let div0;
    	let t6;
    	let t7;
    	let div2;
    	let div1;
    	let t8;
    	let div6;
    	let div4;
    	let t9;
    	let t10;
    	let div5;
    	let t11;
    	let t12;
    	let img1;
    	let img1_src_value;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div7 = element("div");
    			div3 = element("div");
    			h2 = element("h2");
    			t0 = text("Generate API key");
    			t1 = space();
    			input = element("input");
    			t2 = space();
    			button0 = element("button");
    			t3 = text("Generate");
    			t4 = space();
    			button1 = element("button");
    			img0 = element("img");
    			t5 = space();
    			div0 = element("div");
    			t6 = text("Copied!");
    			t7 = space();
    			div2 = element("div");
    			div1 = element("div");
    			t8 = space();
    			div6 = element("div");
    			div4 = element("div");
    			t9 = text("Keep your API key safe and secure.");
    			t10 = space();
    			div5 = element("div");
    			t11 = text("API Analytics");
    			t12 = space();
    			img1 = element("img");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div7 = claim_element(nodes, "DIV", { class: true });
    			var div7_nodes = children(div7);
    			div3 = claim_element(div7_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			h2 = claim_element(div3_nodes, "H2", { class: true });
    			var h2_nodes = children(h2);
    			t0 = claim_text(h2_nodes, "Generate API key");
    			h2_nodes.forEach(detach_dev);
    			t1 = claim_space(div3_nodes);
    			input = claim_element(div3_nodes, "INPUT", { type: true, class: true });
    			t2 = claim_space(div3_nodes);
    			button0 = claim_element(div3_nodes, "BUTTON", { id: true, class: true });
    			var button0_nodes = children(button0);
    			t3 = claim_text(button0_nodes, "Generate");
    			button0_nodes.forEach(detach_dev);
    			t4 = claim_space(div3_nodes);
    			button1 = claim_element(div3_nodes, "BUTTON", { id: true, class: true });
    			var button1_nodes = children(button1);
    			img0 = claim_element(button1_nodes, "IMG", { class: true, src: true, alt: true });
    			button1_nodes.forEach(detach_dev);
    			t5 = claim_space(div3_nodes);
    			div0 = claim_element(div3_nodes, "DIV", { id: true, class: true });
    			var div0_nodes = children(div0);
    			t6 = claim_text(div0_nodes, "Copied!");
    			div0_nodes.forEach(detach_dev);
    			t7 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			div1 = claim_element(div2_nodes, "DIV", { class: true, style: true });
    			children(div1).forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			t8 = claim_space(div7_nodes);
    			div6 = claim_element(div7_nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div4 = claim_element(div6_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			t9 = claim_text(div4_nodes, "Keep your API key safe and secure.");
    			div4_nodes.forEach(detach_dev);
    			t10 = claim_space(div6_nodes);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			t11 = claim_text(div5_nodes, "API Analytics");
    			div5_nodes.forEach(detach_dev);
    			t12 = claim_space(div6_nodes);
    			img1 = claim_element(div6_nodes, "IMG", { class: true, src: true, alt: true });
    			div6_nodes.forEach(detach_dev);
    			div7_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(h2, "class", "svelte-r5bis2");
    			add_location(h2, file$o, 29, 4, 851);
    			attr_dev(input, "type", "text");
    			input.readOnly = true;
    			attr_dev(input, "class", "svelte-r5bis2");
    			add_location(input, file$o, 30, 4, 882);
    			attr_dev(button0, "id", "generateBtn");
    			attr_dev(button0, "class", "svelte-r5bis2");
    			add_location(button0, file$o, 31, 4, 938);
    			attr_dev(img0, "class", "copy-icon svelte-r5bis2");
    			if (!src_url_equal(img0.src, img0_src_value = "img/copy.png")) attr_dev(img0, "src", img0_src_value);
    			attr_dev(img0, "alt", "");
    			add_location(img0, file$o, 35, 7, 1121);
    			attr_dev(button1, "id", "copyBtn");
    			attr_dev(button1, "class", "svelte-r5bis2");
    			add_location(button1, file$o, 34, 4, 1045);
    			attr_dev(div0, "id", "copied");
    			attr_dev(div0, "class", "svelte-r5bis2");
    			add_location(div0, file$o, 37, 4, 1193);
    			attr_dev(div1, "class", "loader");
    			set_style(div1, "display", /*loading*/ ctx[0] ? 'initial' : 'none');
    			add_location(div1, file$o, 40, 6, 1291);
    			attr_dev(div2, "class", "spinner svelte-r5bis2");
    			add_location(div2, file$o, 39, 4, 1262);
    			attr_dev(div3, "class", "content svelte-r5bis2");
    			add_location(div3, file$o, 28, 2, 824);
    			attr_dev(div4, "class", "keep-secure svelte-r5bis2");
    			add_location(div4, file$o, 44, 4, 1414);
    			attr_dev(div5, "class", "highlight logo svelte-r5bis2");
    			add_location(div5, file$o, 45, 4, 1485);
    			attr_dev(img1, "class", "footer-logo");
    			if (!src_url_equal(img1.src, img1_src_value = "img/logo.png")) attr_dev(img1, "src", img1_src_value);
    			attr_dev(img1, "alt", "");
    			add_location(img1, file$o, 46, 4, 1538);
    			attr_dev(div6, "class", "details svelte-r5bis2");
    			add_location(div6, file$o, 43, 2, 1387);
    			attr_dev(div7, "class", "generate svelte-r5bis2");
    			add_location(div7, file$o, 27, 0, 798);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div7, anchor);
    			append_hydration_dev(div7, div3);
    			append_hydration_dev(div3, h2);
    			append_hydration_dev(h2, t0);
    			append_hydration_dev(div3, t1);
    			append_hydration_dev(div3, input);
    			set_input_value(input, /*apiKey*/ ctx[1]);
    			append_hydration_dev(div3, t2);
    			append_hydration_dev(div3, button0);
    			append_hydration_dev(button0, t3);
    			/*button0_binding*/ ctx[8](button0);
    			append_hydration_dev(div3, t4);
    			append_hydration_dev(div3, button1);
    			append_hydration_dev(button1, img0);
    			/*button1_binding*/ ctx[9](button1);
    			append_hydration_dev(div3, t5);
    			append_hydration_dev(div3, div0);
    			append_hydration_dev(div0, t6);
    			/*div0_binding*/ ctx[10](div0);
    			append_hydration_dev(div3, t7);
    			append_hydration_dev(div3, div2);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div7, t8);
    			append_hydration_dev(div7, div6);
    			append_hydration_dev(div6, div4);
    			append_hydration_dev(div4, t9);
    			append_hydration_dev(div6, t10);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, t11);
    			append_hydration_dev(div6, t12);
    			append_hydration_dev(div6, img1);

    			if (!mounted) {
    				dispose = [
    					listen_dev(input, "input", /*input_input_handler*/ ctx[7]),
    					listen_dev(button0, "click", /*genAPIKey*/ ctx[5], false, false, false),
    					listen_dev(button1, "click", /*copyToClipboard*/ ctx[6], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*apiKey*/ 2 && input.value !== /*apiKey*/ ctx[1]) {
    				set_input_value(input, /*apiKey*/ ctx[1]);
    			}

    			if (dirty & /*loading*/ 1) {
    				set_style(div1, "display", /*loading*/ ctx[0] ? 'initial' : 'none');
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div7);
    			/*button0_binding*/ ctx[8](null);
    			/*button1_binding*/ ctx[9](null);
    			/*div0_binding*/ ctx[10](null);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$p.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$p($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Generate', slots, []);
    	let loading = false;
    	let generatedKey = false;
    	let apiKey = "";
    	let generateBtn;
    	let copyBtn;
    	let copiedNotification;

    	async function genAPIKey() {
    		if (!generatedKey) {
    			$$invalidate(0, loading = true);
    			const response = await fetch("https://api-analytics-server.vercel.app/api/generate-api-key");

    			if (response.status == 200) {
    				const data = await response.json();
    				generatedKey = true;
    				$$invalidate(1, apiKey = data.value);
    				$$invalidate(2, generateBtn.style.display = "none", generateBtn);
    				$$invalidate(3, copyBtn.style.display = "grid", copyBtn);
    			}

    			$$invalidate(0, loading = false);
    		}
    	}

    	function copyToClipboard() {
    		navigator.clipboard.writeText(apiKey);
    		$$invalidate(3, copyBtn.value = "Copied", copyBtn);
    		$$invalidate(4, copiedNotification.style.visibility = "visible", copiedNotification);
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Generate> was created with unknown prop '${key}'`);
    	});

    	function input_input_handler() {
    		apiKey = this.value;
    		$$invalidate(1, apiKey);
    	}

    	function button0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			generateBtn = $$value;
    			$$invalidate(2, generateBtn);
    		});
    	}

    	function button1_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			copyBtn = $$value;
    			$$invalidate(3, copyBtn);
    		});
    	}

    	function div0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			copiedNotification = $$value;
    			$$invalidate(4, copiedNotification);
    		});
    	}

    	$$self.$capture_state = () => ({
    		loading,
    		generatedKey,
    		apiKey,
    		generateBtn,
    		copyBtn,
    		copiedNotification,
    		genAPIKey,
    		copyToClipboard
    	});

    	$$self.$inject_state = $$props => {
    		if ('loading' in $$props) $$invalidate(0, loading = $$props.loading);
    		if ('generatedKey' in $$props) generatedKey = $$props.generatedKey;
    		if ('apiKey' in $$props) $$invalidate(1, apiKey = $$props.apiKey);
    		if ('generateBtn' in $$props) $$invalidate(2, generateBtn = $$props.generateBtn);
    		if ('copyBtn' in $$props) $$invalidate(3, copyBtn = $$props.copyBtn);
    		if ('copiedNotification' in $$props) $$invalidate(4, copiedNotification = $$props.copiedNotification);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		loading,
    		apiKey,
    		generateBtn,
    		copyBtn,
    		copiedNotification,
    		genAPIKey,
    		copyToClipboard,
    		input_input_handler,
    		button0_binding,
    		button1_binding,
    		div0_binding
    	];
    }

    class Generate extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$p, create_fragment$p, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Generate",
    			options,
    			id: create_fragment$p.name
    		});
    	}
    }

    /* src\routes\SignIn.svelte generated by Svelte v3.53.1 */

    const { console: console_1$2 } = globals;
    const file$n = "src\\routes\\SignIn.svelte";

    function create_fragment$o(ctx) {
    	let div6;
    	let div2;
    	let h2;
    	let t0;
    	let t1;
    	let input;
    	let t2;
    	let button;
    	let t3;
    	let t4;
    	let div1;
    	let div0;
    	let t5;
    	let div5;
    	let div3;
    	let t6;
    	let t7;
    	let div4;
    	let t8;
    	let t9;
    	let img;
    	let img_src_value;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div6 = element("div");
    			div2 = element("div");
    			h2 = element("h2");
    			t0 = text("Dashboard");
    			t1 = space();
    			input = element("input");
    			t2 = space();
    			button = element("button");
    			t3 = text("Load");
    			t4 = space();
    			div1 = element("div");
    			div0 = element("div");
    			t5 = space();
    			div5 = element("div");
    			div3 = element("div");
    			t6 = text("Keep your API key safe and secure.");
    			t7 = space();
    			div4 = element("div");
    			t8 = text("API Analytics");
    			t9 = space();
    			img = element("img");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div6 = claim_element(nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div2 = claim_element(div6_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			h2 = claim_element(div2_nodes, "H2", { class: true });
    			var h2_nodes = children(h2);
    			t0 = claim_text(h2_nodes, "Dashboard");
    			h2_nodes.forEach(detach_dev);
    			t1 = claim_space(div2_nodes);

    			input = claim_element(div2_nodes, "INPUT", {
    				type: true,
    				placeholder: true,
    				class: true
    			});

    			t2 = claim_space(div2_nodes);
    			button = claim_element(div2_nodes, "BUTTON", { id: true, class: true });
    			var button_nodes = children(button);
    			t3 = claim_text(button_nodes, "Load");
    			button_nodes.forEach(detach_dev);
    			t4 = claim_space(div2_nodes);
    			div1 = claim_element(div2_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true, style: true });
    			children(div0).forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			t5 = claim_space(div6_nodes);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			div3 = claim_element(div5_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			t6 = claim_text(div3_nodes, "Keep your API key safe and secure.");
    			div3_nodes.forEach(detach_dev);
    			t7 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			t8 = claim_text(div4_nodes, "API Analytics");
    			div4_nodes.forEach(detach_dev);
    			t9 = claim_space(div5_nodes);
    			img = claim_element(div5_nodes, "IMG", { class: true, src: true, alt: true });
    			div5_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(h2, "class", "svelte-1kmr1cu");
    			add_location(h2, file$n, 17, 4, 517);
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "Enter API key");
    			attr_dev(input, "class", "svelte-1kmr1cu");
    			add_location(input, file$n, 18, 4, 541);
    			attr_dev(button, "id", "generateBtn");
    			attr_dev(button, "class", "svelte-1kmr1cu");
    			add_location(button, file$n, 19, 4, 615);
    			attr_dev(div0, "class", "loader");
    			set_style(div0, "display", /*loading*/ ctx[1] ? 'initial' : 'none');
    			add_location(div0, file$n, 21, 6, 709);
    			attr_dev(div1, "class", "spinner");
    			add_location(div1, file$n, 20, 4, 680);
    			attr_dev(div2, "class", "content svelte-1kmr1cu");
    			add_location(div2, file$n, 16, 2, 490);
    			attr_dev(div3, "class", "keep-secure svelte-1kmr1cu");
    			add_location(div3, file$n, 25, 4, 832);
    			attr_dev(div4, "class", "highlight logo svelte-1kmr1cu");
    			add_location(div4, file$n, 26, 4, 903);
    			attr_dev(img, "class", "footer-logo");
    			if (!src_url_equal(img.src, img_src_value = "img/logo.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			add_location(img, file$n, 27, 4, 956);
    			attr_dev(div5, "class", "details svelte-1kmr1cu");
    			add_location(div5, file$n, 24, 2, 805);
    			attr_dev(div6, "class", "generate svelte-1kmr1cu");
    			add_location(div6, file$n, 15, 0, 464);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div6, anchor);
    			append_hydration_dev(div6, div2);
    			append_hydration_dev(div2, h2);
    			append_hydration_dev(h2, t0);
    			append_hydration_dev(div2, t1);
    			append_hydration_dev(div2, input);
    			set_input_value(input, /*apiKey*/ ctx[0]);
    			append_hydration_dev(div2, t2);
    			append_hydration_dev(div2, button);
    			append_hydration_dev(button, t3);
    			append_hydration_dev(div2, t4);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div6, t5);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, div3);
    			append_hydration_dev(div3, t6);
    			append_hydration_dev(div5, t7);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, t8);
    			append_hydration_dev(div5, t9);
    			append_hydration_dev(div5, img);

    			if (!mounted) {
    				dispose = [
    					listen_dev(input, "input", /*input_input_handler*/ ctx[3]),
    					listen_dev(button, "click", /*genAPIKey*/ ctx[2], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*apiKey*/ 1 && input.value !== /*apiKey*/ ctx[0]) {
    				set_input_value(input, /*apiKey*/ ctx[0]);
    			}

    			if (dirty & /*loading*/ 2) {
    				set_style(div0, "display", /*loading*/ ctx[1] ? 'initial' : 'none');
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div6);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$o.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$o($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('SignIn', slots, []);
    	let apiKey = "";
    	let loading = false;

    	async function genAPIKey() {
    		$$invalidate(1, loading = true);

    		// Fetch page ID
    		const response = await fetch(`https://api-analytics-server.vercel.app/api/user-id/${apiKey}`);

    		console.log(response);

    		if (response.status == 200) {
    			const data = await response.json();
    			window.location.href = `/dashboard/${data.value.replaceAll("-", "")}`;
    		}

    		$$invalidate(1, loading = false);
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console_1$2.warn(`<SignIn> was created with unknown prop '${key}'`);
    	});

    	function input_input_handler() {
    		apiKey = this.value;
    		$$invalidate(0, apiKey);
    	}

    	$$self.$capture_state = () => ({ apiKey, loading, genAPIKey });

    	$$self.$inject_state = $$props => {
    		if ('apiKey' in $$props) $$invalidate(0, apiKey = $$props.apiKey);
    		if ('loading' in $$props) $$invalidate(1, loading = $$props.loading);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [apiKey, loading, genAPIKey, input_input_handler];
    }

    class SignIn extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$o, create_fragment$o, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "SignIn",
    			options,
    			id: create_fragment$o.name
    		});
    	}
    }

    /* src\components\dashboard\Requests.svelte generated by Svelte v3.53.1 */
    const file$m = "src\\components\\dashboard\\Requests.svelte";

    // (20:2) {#if percentageChange != null}
    function create_if_block$9(ctx) {
    	let div;
    	let t0;
    	let t1_value = (/*percentageChange*/ ctx[1] > 0 ? "+" : "") + "";
    	let t1;
    	let t2_value = /*percentageChange*/ ctx[1].toFixed(1) + "";
    	let t2;
    	let t3;

    	const block = {
    		c: function create() {
    			div = element("div");
    			t0 = text("(");
    			t1 = text(t1_value);
    			t2 = text(t2_value);
    			t3 = text("%)");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true });
    			var div_nodes = children(div);
    			t0 = claim_text(div_nodes, "(");
    			t1 = claim_text(div_nodes, t1_value);
    			t2 = claim_text(div_nodes, t2_value);
    			t3 = claim_text(div_nodes, "%)");
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "percentage-change svelte-1mgix01");
    			toggle_class(div, "positive", /*percentageChange*/ ctx[1] > 0);
    			toggle_class(div, "negative", /*percentageChange*/ ctx[1] < 0);
    			add_location(div, file$m, 20, 4, 503);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    			append_hydration_dev(div, t0);
    			append_hydration_dev(div, t1);
    			append_hydration_dev(div, t2);
    			append_hydration_dev(div, t3);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*percentageChange*/ 2 && t1_value !== (t1_value = (/*percentageChange*/ ctx[1] > 0 ? "+" : "") + "")) set_data_dev(t1, t1_value);
    			if (dirty & /*percentageChange*/ 2 && t2_value !== (t2_value = /*percentageChange*/ ctx[1].toFixed(1) + "")) set_data_dev(t2, t2_value);

    			if (dirty & /*percentageChange*/ 2) {
    				toggle_class(div, "positive", /*percentageChange*/ ctx[1] > 0);
    			}

    			if (dirty & /*percentageChange*/ 2) {
    				toggle_class(div, "negative", /*percentageChange*/ ctx[1] < 0);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$9.name,
    		type: "if",
    		source: "(20:2) {#if percentageChange != null}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$n(ctx) {
    	let div2;
    	let t0;
    	let div0;
    	let t1;
    	let t2;
    	let div1;
    	let t3_value = /*data*/ ctx[0].length.toLocaleString() + "";
    	let t3;
    	let if_block = /*percentageChange*/ ctx[1] != null && create_if_block$9(ctx);

    	const block = {
    		c: function create() {
    			div2 = element("div");
    			if (if_block) if_block.c();
    			t0 = space();
    			div0 = element("div");
    			t1 = text("Requests");
    			t2 = space();
    			div1 = element("div");
    			t3 = text(t3_value);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div2 = claim_element(nodes, "DIV", { class: true, title: true });
    			var div2_nodes = children(div2);
    			if (if_block) if_block.l(div2_nodes);
    			t0 = claim_space(div2_nodes);
    			div0 = claim_element(div2_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t1 = claim_text(div0_nodes, "Requests");
    			div0_nodes.forEach(detach_dev);
    			t2 = claim_space(div2_nodes);
    			div1 = claim_element(div2_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			t3 = claim_text(div1_nodes, t3_value);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$m, 28, 2, 735);
    			attr_dev(div1, "class", "value svelte-1mgix01");
    			add_location(div1, file$m, 29, 2, 777);
    			attr_dev(div2, "class", "card svelte-1mgix01");
    			attr_dev(div2, "title", "Total");
    			add_location(div2, file$m, 18, 0, 431);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div2, anchor);
    			if (if_block) if_block.m(div2, null);
    			append_hydration_dev(div2, t0);
    			append_hydration_dev(div2, div0);
    			append_hydration_dev(div0, t1);
    			append_hydration_dev(div2, t2);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div1, t3);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*percentageChange*/ ctx[1] != null) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$9(ctx);
    					if_block.c();
    					if_block.m(div2, t0);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}

    			if (dirty & /*data*/ 1 && t3_value !== (t3_value = /*data*/ ctx[0].length.toLocaleString() + "")) set_data_dev(t3, t3_value);
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div2);
    			if (if_block) if_block.d();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$n.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$n($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Requests', slots, []);

    	function setPercentageChange() {
    		if (prevData.length == 0) {
    			$$invalidate(1, percentageChange = null);
    		} else {
    			$$invalidate(1, percentageChange = data.length / prevData.length * 100 - 100);
    		}
    	}

    	let percentageChange;
    	let mounted = false;

    	onMount(() => {
    		$$invalidate(3, mounted = true);
    	});

    	let { data, prevData } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Requests> was created without expected prop 'data'");
    		}

    		if (prevData === undefined && !('prevData' in $$props || $$self.$$.bound[$$self.$$.props['prevData']])) {
    			console.warn("<Requests> was created without expected prop 'prevData'");
    		}
    	});

    	const writable_props = ['data', 'prevData'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Requests> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    		if ('prevData' in $$props) $$invalidate(2, prevData = $$props.prevData);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		setPercentageChange,
    		percentageChange,
    		mounted,
    		data,
    		prevData
    	});

    	$$self.$inject_state = $$props => {
    		if ('percentageChange' in $$props) $$invalidate(1, percentageChange = $$props.percentageChange);
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    		if ('prevData' in $$props) $$invalidate(2, prevData = $$props.prevData);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 9) {
    			data && mounted && setPercentageChange();
    		}
    	};

    	return [data, percentageChange, prevData, mounted];
    }

    class Requests extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$n, create_fragment$n, safe_not_equal, { data: 0, prevData: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Requests",
    			options,
    			id: create_fragment$n.name
    		});
    	}

    	get data() {
    		throw new Error("<Requests>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Requests>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get prevData() {
    		throw new Error("<Requests>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set prevData(value) {
    		throw new Error("<Requests>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\Welcome.svelte generated by Svelte v3.53.1 */

    const file$l = "src\\components\\dashboard\\Welcome.svelte";

    function create_fragment$m(ctx) {
    	let div;
    	let img;
    	let img_src_value;

    	const block = {
    		c: function create() {
    			div = element("div");
    			img = element("img");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true, title: true });
    			var div_nodes = children(div);
    			img = claim_element(div_nodes, "IMG", { src: true, alt: true, class: true });
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			if (!src_url_equal(img.src, img_src_value = "../img/logo.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			attr_dev(img, "class", "svelte-1w2ck9z");
    			add_location(img, file$l, 1, 2, 44);
    			attr_dev(div, "class", "card svelte-1w2ck9z");
    			attr_dev(div, "title", "API Analytics");
    			add_location(div, file$l, 0, 0, 0);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    			append_hydration_dev(div, img);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$m.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$m($$self, $$props) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Welcome', slots, []);
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Welcome> was created with unknown prop '${key}'`);
    	});

    	return [];
    }

    class Welcome extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$m, create_fragment$m, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Welcome",
    			options,
    			id: create_fragment$m.name
    		});
    	}
    }

    /* src\components\dashboard\RequestsPerHour.svelte generated by Svelte v3.53.1 */
    const file$k = "src\\components\\dashboard\\RequestsPerHour.svelte";

    // (53:2) {#if requestsPerHour != undefined}
    function create_if_block$8(ctx) {
    	let div;
    	let t;

    	const block = {
    		c: function create() {
    			div = element("div");
    			t = text(/*requestsPerHour*/ ctx[0]);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true });
    			var div_nodes = children(div);
    			t = claim_text(div_nodes, /*requestsPerHour*/ ctx[0]);
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "value svelte-r5qgcj");
    			add_location(div, file$k, 53, 4, 1210);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    			append_hydration_dev(div, t);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*requestsPerHour*/ 1) set_data_dev(t, /*requestsPerHour*/ ctx[0]);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$8.name,
    		type: "if",
    		source: "(53:2) {#if requestsPerHour != undefined}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$l(ctx) {
    	let div1;
    	let div0;
    	let t0;
    	let span;
    	let t1;
    	let t2;
    	let if_block = /*requestsPerHour*/ ctx[0] != undefined && create_if_block$8(ctx);

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			t0 = text("Requests ");
    			span = element("span");
    			t1 = text("/ hour");
    			t2 = space();
    			if (if_block) if_block.c();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true, title: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Requests ");
    			span = claim_element(div0_nodes, "SPAN", { class: true });
    			var span_nodes = children(span);
    			t1 = claim_text(span_nodes, "/ hour");
    			span_nodes.forEach(detach_dev);
    			div0_nodes.forEach(detach_dev);
    			t2 = claim_space(div1_nodes);
    			if (if_block) if_block.l(div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(span, "class", "per-hour svelte-r5qgcj");
    			add_location(span, file$k, 50, 13, 1120);
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$k, 49, 2, 1081);
    			attr_dev(div1, "class", "card svelte-r5qgcj");
    			attr_dev(div1, "title", "Last week");
    			add_location(div1, file$k, 48, 0, 1041);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div0, span);
    			append_hydration_dev(span, t1);
    			append_hydration_dev(div1, t2);
    			if (if_block) if_block.m(div1, null);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*requestsPerHour*/ ctx[0] != undefined) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$8(ctx);
    					if_block.c();
    					if_block.m(div1, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			if (if_block) if_block.d();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$l.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function periodToDays$4(period) {
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

    function instance$l($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('RequestsPerHour', slots, []);

    	function build() {
    		let totalRequests = 0;

    		for (let i = 0; i < data.length; i++) {
    			totalRequests++;
    		}

    		if (totalRequests > 0) {
    			let days = periodToDays$4(period);

    			if (days != null) {
    				$$invalidate(0, requestsPerHour = (totalRequests / (24 * days)).toFixed(2));
    			}
    		} else {
    			$$invalidate(0, requestsPerHour = "0");
    		}
    	}

    	let mounted = false;
    	let requestsPerHour;

    	onMount(() => {
    		$$invalidate(3, mounted = true);
    	});

    	let { data, period } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<RequestsPerHour> was created without expected prop 'data'");
    		}

    		if (period === undefined && !('period' in $$props || $$self.$$.bound[$$self.$$.props['period']])) {
    			console.warn("<RequestsPerHour> was created without expected prop 'period'");
    		}
    	});

    	const writable_props = ['data', 'period'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<RequestsPerHour> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		periodToDays: periodToDays$4,
    		build,
    		mounted,
    		requestsPerHour,
    		data,
    		period
    	});

    	$$self.$inject_state = $$props => {
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('requestsPerHour' in $$props) $$invalidate(0, requestsPerHour = $$props.requestsPerHour);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 10) {
    			data && mounted && build();
    		}
    	};

    	return [requestsPerHour, data, period, mounted];
    }

    class RequestsPerHour extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$l, create_fragment$l, safe_not_equal, { data: 1, period: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "RequestsPerHour",
    			options,
    			id: create_fragment$l.name
    		});
    	}

    	get data() {
    		throw new Error("<RequestsPerHour>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<RequestsPerHour>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get period() {
    		throw new Error("<RequestsPerHour>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set period(value) {
    		throw new Error("<RequestsPerHour>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\ResponseTimes.svelte generated by Svelte v3.53.1 */
    const file$j = "src\\components\\dashboard\\ResponseTimes.svelte";

    function create_fragment$k(ctx) {
    	let div14;
    	let div0;
    	let t0;
    	let span;
    	let t1;
    	let t2;
    	let div4;
    	let div1;
    	let t3;
    	let t4;
    	let div2;
    	let t5;
    	let t6;
    	let div3;
    	let t7;
    	let t8;
    	let div8;
    	let div5;
    	let t9;
    	let t10;
    	let div6;
    	let t11;
    	let t12;
    	let div7;
    	let t13;
    	let t14;
    	let div13;
    	let div9;
    	let t15;
    	let div10;
    	let t16;
    	let div11;
    	let t17;
    	let div12;

    	const block = {
    		c: function create() {
    			div14 = element("div");
    			div0 = element("div");
    			t0 = text("Response times ");
    			span = element("span");
    			t1 = text("(ms)");
    			t2 = space();
    			div4 = element("div");
    			div1 = element("div");
    			t3 = text(/*LQ*/ ctx[1]);
    			t4 = space();
    			div2 = element("div");
    			t5 = text(/*median*/ ctx[0]);
    			t6 = space();
    			div3 = element("div");
    			t7 = text(/*UQ*/ ctx[2]);
    			t8 = space();
    			div8 = element("div");
    			div5 = element("div");
    			t9 = text("25%");
    			t10 = space();
    			div6 = element("div");
    			t11 = text("Median");
    			t12 = space();
    			div7 = element("div");
    			t13 = text("75%");
    			t14 = space();
    			div13 = element("div");
    			div9 = element("div");
    			t15 = space();
    			div10 = element("div");
    			t16 = space();
    			div11 = element("div");
    			t17 = space();
    			div12 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div14 = claim_element(nodes, "DIV", { class: true });
    			var div14_nodes = children(div14);
    			div0 = claim_element(div14_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Response times ");
    			span = claim_element(div0_nodes, "SPAN", { class: true });
    			var span_nodes = children(span);
    			t1 = claim_text(span_nodes, "(ms)");
    			span_nodes.forEach(detach_dev);
    			div0_nodes.forEach(detach_dev);
    			t2 = claim_space(div14_nodes);
    			div4 = claim_element(div14_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			div1 = claim_element(div4_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			t3 = claim_text(div1_nodes, /*LQ*/ ctx[1]);
    			div1_nodes.forEach(detach_dev);
    			t4 = claim_space(div4_nodes);
    			div2 = claim_element(div4_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			t5 = claim_text(div2_nodes, /*median*/ ctx[0]);
    			div2_nodes.forEach(detach_dev);
    			t6 = claim_space(div4_nodes);
    			div3 = claim_element(div4_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			t7 = claim_text(div3_nodes, /*UQ*/ ctx[2]);
    			div3_nodes.forEach(detach_dev);
    			div4_nodes.forEach(detach_dev);
    			t8 = claim_space(div14_nodes);
    			div8 = claim_element(div14_nodes, "DIV", { class: true });
    			var div8_nodes = children(div8);
    			div5 = claim_element(div8_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			t9 = claim_text(div5_nodes, "25%");
    			div5_nodes.forEach(detach_dev);
    			t10 = claim_space(div8_nodes);
    			div6 = claim_element(div8_nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			t11 = claim_text(div6_nodes, "Median");
    			div6_nodes.forEach(detach_dev);
    			t12 = claim_space(div8_nodes);
    			div7 = claim_element(div8_nodes, "DIV", { class: true });
    			var div7_nodes = children(div7);
    			t13 = claim_text(div7_nodes, "75%");
    			div7_nodes.forEach(detach_dev);
    			div8_nodes.forEach(detach_dev);
    			t14 = claim_space(div14_nodes);
    			div13 = claim_element(div14_nodes, "DIV", { class: true });
    			var div13_nodes = children(div13);
    			div9 = claim_element(div13_nodes, "DIV", { class: true });
    			children(div9).forEach(detach_dev);
    			t15 = claim_space(div13_nodes);
    			div10 = claim_element(div13_nodes, "DIV", { class: true });
    			children(div10).forEach(detach_dev);
    			t16 = claim_space(div13_nodes);
    			div11 = claim_element(div13_nodes, "DIV", { class: true });
    			children(div11).forEach(detach_dev);
    			t17 = claim_space(div13_nodes);
    			div12 = claim_element(div13_nodes, "DIV", { class: true });
    			children(div12).forEach(detach_dev);
    			div13_nodes.forEach(detach_dev);
    			div14_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(span, "class", "milliseconds svelte-kx5j01");
    			add_location(span, file$j, 67, 19, 1815);
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$j, 66, 2, 1770);
    			attr_dev(div1, "class", "value lower-quartile svelte-kx5j01");
    			add_location(div1, file$j, 70, 4, 1893);
    			attr_dev(div2, "class", "value median svelte-kx5j01");
    			add_location(div2, file$j, 71, 4, 1943);
    			attr_dev(div3, "class", "value upper-quartile svelte-kx5j01");
    			add_location(div3, file$j, 72, 4, 1989);
    			attr_dev(div4, "class", "values svelte-kx5j01");
    			add_location(div4, file$j, 69, 2, 1867);
    			attr_dev(div5, "class", "label svelte-kx5j01");
    			add_location(div5, file$j, 75, 4, 2073);
    			attr_dev(div6, "class", "label svelte-kx5j01");
    			add_location(div6, file$j, 76, 4, 2107);
    			attr_dev(div7, "class", "label svelte-kx5j01");
    			add_location(div7, file$j, 77, 4, 2144);
    			attr_dev(div8, "class", "labels svelte-kx5j01");
    			add_location(div8, file$j, 74, 2, 2047);
    			attr_dev(div9, "class", "bar-green svelte-kx5j01");
    			add_location(div9, file$j, 80, 4, 2209);
    			attr_dev(div10, "class", "bar-yellow svelte-kx5j01");
    			add_location(div10, file$j, 81, 4, 2240);
    			attr_dev(div11, "class", "bar-red svelte-kx5j01");
    			add_location(div11, file$j, 82, 4, 2272);
    			attr_dev(div12, "class", "marker svelte-kx5j01");
    			add_location(div12, file$j, 83, 4, 2301);
    			attr_dev(div13, "class", "bar svelte-kx5j01");
    			add_location(div13, file$j, 79, 2, 2186);
    			attr_dev(div14, "class", "card");
    			add_location(div14, file$j, 65, 0, 1748);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div14, anchor);
    			append_hydration_dev(div14, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div0, span);
    			append_hydration_dev(span, t1);
    			append_hydration_dev(div14, t2);
    			append_hydration_dev(div14, div4);
    			append_hydration_dev(div4, div1);
    			append_hydration_dev(div1, t3);
    			append_hydration_dev(div4, t4);
    			append_hydration_dev(div4, div2);
    			append_hydration_dev(div2, t5);
    			append_hydration_dev(div4, t6);
    			append_hydration_dev(div4, div3);
    			append_hydration_dev(div3, t7);
    			append_hydration_dev(div14, t8);
    			append_hydration_dev(div14, div8);
    			append_hydration_dev(div8, div5);
    			append_hydration_dev(div5, t9);
    			append_hydration_dev(div8, t10);
    			append_hydration_dev(div8, div6);
    			append_hydration_dev(div6, t11);
    			append_hydration_dev(div8, t12);
    			append_hydration_dev(div8, div7);
    			append_hydration_dev(div7, t13);
    			append_hydration_dev(div14, t14);
    			append_hydration_dev(div14, div13);
    			append_hydration_dev(div13, div9);
    			append_hydration_dev(div13, t15);
    			append_hydration_dev(div13, div10);
    			append_hydration_dev(div13, t16);
    			append_hydration_dev(div13, div11);
    			append_hydration_dev(div13, t17);
    			append_hydration_dev(div13, div12);
    			/*div12_binding*/ ctx[6](div12);
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*LQ*/ 2) set_data_dev(t3, /*LQ*/ ctx[1]);
    			if (dirty & /*median*/ 1) set_data_dev(t5, /*median*/ ctx[0]);
    			if (dirty & /*UQ*/ 4) set_data_dev(t7, /*UQ*/ ctx[2]);
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div14);
    			/*div12_binding*/ ctx[6](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$k.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$k($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('ResponseTimes', slots, []);
    	const asc = arr => arr.sort((a, b) => a - b);
    	const sum = arr => arr.reduce((a, b) => a + b, 0);
    	const mean = arr => sum(arr) / arr.length;

    	// sample standard deviation
    	function std(arr) {
    		const mu = mean(arr);
    		const diffArr = arr.map(a => (a - mu) ** 2);
    		return Math.sqrt(sum(diffArr) / (arr.length - 1));
    	}

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
    		$$invalidate(3, marker.style.left = `${position}%`, marker);
    	}

    	function build() {
    		let responseTimes = [];

    		for (let i = 0; i < data.length; i++) {
    			responseTimes.push(data[i].response_time);
    		}

    		$$invalidate(1, LQ = quantile(responseTimes, 0.25));
    		$$invalidate(0, median = quantile(responseTimes, 0.5));
    		$$invalidate(2, UQ = quantile(responseTimes, 0.75));
    		setMarkerPosition(median);
    	}

    	let median;
    	let LQ;
    	let UQ;
    	let marker;
    	let mounted = false;

    	onMount(() => {
    		$$invalidate(5, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<ResponseTimes> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<ResponseTimes> was created with unknown prop '${key}'`);
    	});

    	function div12_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			marker = $$value;
    			$$invalidate(3, marker);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(4, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		asc,
    		sum,
    		mean,
    		std,
    		quantile,
    		markerPosition,
    		setMarkerPosition,
    		build,
    		median,
    		LQ,
    		UQ,
    		marker,
    		mounted,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('median' in $$props) $$invalidate(0, median = $$props.median);
    		if ('LQ' in $$props) $$invalidate(1, LQ = $$props.LQ);
    		if ('UQ' in $$props) $$invalidate(2, UQ = $$props.UQ);
    		if ('marker' in $$props) $$invalidate(3, marker = $$props.marker);
    		if ('mounted' in $$props) $$invalidate(5, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(4, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 48) {
    			data && mounted && build();
    		}
    	};

    	return [median, LQ, UQ, marker, data, mounted, div12_binding];
    }

    class ResponseTimes extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$k, create_fragment$k, safe_not_equal, { data: 4 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "ResponseTimes",
    			options,
    			id: create_fragment$k.name
    		});
    	}

    	get data() {
    		throw new Error("<ResponseTimes>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<ResponseTimes>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\Endpoints.svelte generated by Svelte v3.53.1 */
    const file$i = "src\\components\\dashboard\\Endpoints.svelte";

    function get_each_context$2(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[8] = list[i];
    	child_ctx[10] = i;
    	return child_ctx;
    }

    // (65:2) {#if endpoints != undefined}
    function create_if_block$7(ctx) {
    	let div;
    	let each_value = /*endpoints*/ ctx[0];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block$2(get_each_context$2(ctx, each_value, i));
    	}

    	const block = {
    		c: function create() {
    			div = element("div");

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true });
    			var div_nodes = children(div);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].l(div_nodes);
    			}

    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "endpoints svelte-1l35be3");
    			add_location(div, file$i, 65, 4, 2010);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(div, null);
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*endpoints, maxCount*/ 3) {
    				each_value = /*endpoints*/ ctx[0];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context$2(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block$2(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(div, null);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			destroy_each(each_blocks, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$7.name,
    		type: "if",
    		source: "(65:2) {#if endpoints != undefined}",
    		ctx
    	});

    	return block;
    }

    // (67:6) {#each endpoints as endpoint, i}
    function create_each_block$2(ctx) {
    	let div6;
    	let div3;
    	let div2;
    	let div0;
    	let t0_value = /*endpoint*/ ctx[8].path + "";
    	let t0;
    	let t1;
    	let div1;
    	let t2_value = /*endpoint*/ ctx[8].count + "";
    	let t2;
    	let div3_title_value;
    	let t3;
    	let div5;
    	let div4;
    	let t4_value = /*endpoint*/ ctx[8].path + "";
    	let t4;
    	let t5;

    	const block = {
    		c: function create() {
    			div6 = element("div");
    			div3 = element("div");
    			div2 = element("div");
    			div0 = element("div");
    			t0 = text(t0_value);
    			t1 = space();
    			div1 = element("div");
    			t2 = text(t2_value);
    			t3 = space();
    			div5 = element("div");
    			div4 = element("div");
    			t4 = text(t4_value);
    			t5 = space();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div6 = claim_element(nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);

    			div3 = claim_element(div6_nodes, "DIV", {
    				class: true,
    				id: true,
    				title: true,
    				style: true
    			});

    			var div3_nodes = children(div3);
    			div2 = claim_element(div3_nodes, "DIV", { class: true, id: true });
    			var div2_nodes = children(div2);
    			div0 = claim_element(div2_nodes, "DIV", { class: true, id: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, t0_value);
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div2_nodes);
    			div1 = claim_element(div2_nodes, "DIV", { class: true, id: true });
    			var div1_nodes = children(div1);
    			t2 = claim_text(div1_nodes, t2_value);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			t3 = claim_space(div6_nodes);
    			div5 = claim_element(div6_nodes, "DIV", { class: true, id: true });
    			var div5_nodes = children(div5);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			t4 = claim_text(div4_nodes, t4_value);
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			t5 = claim_space(div6_nodes);
    			div6_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "path svelte-1l35be3");
    			attr_dev(div0, "id", "endpoint-path-" + /*i*/ ctx[10]);
    			add_location(div0, file$i, 79, 14, 2543);
    			attr_dev(div1, "class", "count svelte-1l35be3");
    			attr_dev(div1, "id", "endpoint-count-" + /*i*/ ctx[10]);
    			add_location(div1, file$i, 82, 14, 2655);
    			attr_dev(div2, "class", "endpoint-label svelte-1l35be3");
    			attr_dev(div2, "id", "endpoint-label-" + /*i*/ ctx[10]);
    			add_location(div2, file$i, 78, 12, 2475);
    			attr_dev(div3, "class", "endpoint svelte-1l35be3");
    			attr_dev(div3, "id", "endpoint-" + /*i*/ ctx[10]);
    			attr_dev(div3, "title", div3_title_value = /*endpoint*/ ctx[8].count);
    			set_style(div3, "width", /*endpoint*/ ctx[8].count / /*maxCount*/ ctx[1] * 100 + "%");

    			set_style(div3, "background", /*endpoint*/ ctx[8].status >= 200 && /*endpoint*/ ctx[8].status <= 299
    			? 'var(--highlight)'
    			: '#e46161');

    			add_location(div3, file$i, 68, 10, 2127);
    			attr_dev(div4, "class", "external-label-path");
    			add_location(div4, file$i, 86, 12, 2836);
    			attr_dev(div5, "class", "external-label svelte-1l35be3");
    			attr_dev(div5, "id", "external-label-" + /*i*/ ctx[10]);
    			add_location(div5, file$i, 85, 10, 2770);
    			attr_dev(div6, "class", "endpoint-container svelte-1l35be3");
    			add_location(div6, file$i, 67, 8, 2083);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div6, anchor);
    			append_hydration_dev(div6, div3);
    			append_hydration_dev(div3, div2);
    			append_hydration_dev(div2, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div2, t1);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div1, t2);
    			append_hydration_dev(div6, t3);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, t4);
    			append_hydration_dev(div6, t5);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*endpoints*/ 1 && t0_value !== (t0_value = /*endpoint*/ ctx[8].path + "")) set_data_dev(t0, t0_value);
    			if (dirty & /*endpoints*/ 1 && t2_value !== (t2_value = /*endpoint*/ ctx[8].count + "")) set_data_dev(t2, t2_value);

    			if (dirty & /*endpoints*/ 1 && div3_title_value !== (div3_title_value = /*endpoint*/ ctx[8].count)) {
    				attr_dev(div3, "title", div3_title_value);
    			}

    			if (dirty & /*endpoints, maxCount*/ 3) {
    				set_style(div3, "width", /*endpoint*/ ctx[8].count / /*maxCount*/ ctx[1] * 100 + "%");
    			}

    			if (dirty & /*endpoints*/ 1) {
    				set_style(div3, "background", /*endpoint*/ ctx[8].status >= 200 && /*endpoint*/ ctx[8].status <= 299
    				? 'var(--highlight)'
    				: '#e46161');
    			}

    			if (dirty & /*endpoints*/ 1 && t4_value !== (t4_value = /*endpoint*/ ctx[8].path + "")) set_data_dev(t4, t4_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div6);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block$2.name,
    		type: "each",
    		source: "(67:6) {#each endpoints as endpoint, i}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$j(ctx) {
    	let div1;
    	let div0;
    	let t0;
    	let t1;
    	let if_block = /*endpoints*/ ctx[0] != undefined && create_if_block$7(ctx);

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			t0 = text("Endpoints");
    			t1 = space();
    			if (if_block) if_block.c();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Endpoints");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div1_nodes);
    			if (if_block) if_block.l(div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$i, 63, 2, 1933);
    			attr_dev(div1, "class", "card svelte-1l35be3");
    			add_location(div1, file$i, 62, 0, 1911);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div1, t1);
    			if (if_block) if_block.m(div1, null);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*endpoints*/ ctx[0] != undefined) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$7(ctx);
    					if_block.c();
    					if_block.m(div1, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			if (if_block) if_block.d();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$j.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$j($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Endpoints', slots, []);
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
    		$$invalidate(1, maxCount = 0);

    		for (let endpointID in freq) {
    			freqArr.push(freq[endpointID]);

    			if (freq[endpointID].count > maxCount) {
    				$$invalidate(1, maxCount = freq[endpointID].count);
    			}
    		}

    		freqArr.sort((a, b) => {
    			return b.count - a.count;
    		});

    		$$invalidate(0, endpoints = freqArr);
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
    		$$invalidate(3, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Endpoints> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Endpoints> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(2, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		methodMap,
    		endpointFreq,
    		build,
    		setEndpointLabelVisibility,
    		setEndpointLabels,
    		endpoints,
    		maxCount,
    		mounted,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('methodMap' in $$props) methodMap = $$props.methodMap;
    		if ('endpoints' in $$props) $$invalidate(0, endpoints = $$props.endpoints);
    		if ('maxCount' in $$props) $$invalidate(1, maxCount = $$props.maxCount);
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(2, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 12) {
    			data && mounted && build();
    		}
    	};

    	return [endpoints, maxCount, data, mounted];
    }

    class Endpoints extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$j, create_fragment$j, safe_not_equal, { data: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Endpoints",
    			options,
    			id: create_fragment$j.name
    		});
    	}

    	get data() {
    		throw new Error("<Endpoints>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Endpoints>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\SuccessRate.svelte generated by Svelte v3.53.1 */
    const file$h = "src\\components\\dashboard\\SuccessRate.svelte";

    // (29:2) {#if successRate != undefined}
    function create_if_block$6(ctx) {
    	let div;
    	let t0_value = /*successRate*/ ctx[0].toFixed(1) + "";
    	let t0;
    	let t1;

    	const block = {
    		c: function create() {
    			div = element("div");
    			t0 = text(t0_value);
    			t1 = text("%");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true, style: true });
    			var div_nodes = children(div);
    			t0 = claim_text(div_nodes, t0_value);
    			t1 = claim_text(div_nodes, "%");
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "value svelte-1vzzb7c");

    			set_style(div, "color", (/*successRate*/ ctx[0] <= 75 ? 'var(--red)' : '') + (/*successRate*/ ctx[0] > 75 && /*successRate*/ ctx[0] < 90
    			? 'var(--yellow)'
    			: '') + (/*successRate*/ ctx[0] >= 90 ? 'var(--highlight)' : ''));

    			add_location(div, file$h, 29, 4, 743);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    			append_hydration_dev(div, t0);
    			append_hydration_dev(div, t1);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*successRate*/ 1 && t0_value !== (t0_value = /*successRate*/ ctx[0].toFixed(1) + "")) set_data_dev(t0, t0_value);

    			if (dirty & /*successRate*/ 1) {
    				set_style(div, "color", (/*successRate*/ ctx[0] <= 75 ? 'var(--red)' : '') + (/*successRate*/ ctx[0] > 75 && /*successRate*/ ctx[0] < 90
    				? 'var(--yellow)'
    				: '') + (/*successRate*/ ctx[0] >= 90 ? 'var(--highlight)' : ''));
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$6.name,
    		type: "if",
    		source: "(29:2) {#if successRate != undefined}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$i(ctx) {
    	let div1;
    	let div0;
    	let t0;
    	let t1;
    	let if_block = /*successRate*/ ctx[0] != undefined && create_if_block$6(ctx);

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			t0 = text("Success rate");
    			t1 = space();
    			if (if_block) if_block.c();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true, title: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Success rate");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div1_nodes);
    			if (if_block) if_block.l(div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$h, 27, 2, 661);
    			attr_dev(div1, "class", "card svelte-1vzzb7c");
    			attr_dev(div1, "title", "Last week");
    			add_location(div1, file$h, 26, 0, 621);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div1, t1);
    			if (if_block) if_block.m(div1, null);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*successRate*/ ctx[0] != undefined) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$6(ctx);
    					if_block.c();
    					if_block.m(div1, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			if (if_block) if_block.d();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$i.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$i($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('SuccessRate', slots, []);

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
    			$$invalidate(0, successRate = successfulRequests / totalRequests * 100);
    		} else {
    			$$invalidate(0, successRate = 100);
    		}
    	}

    	let mounted = false;
    	let successRate;

    	onMount(() => {
    		$$invalidate(2, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<SuccessRate> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<SuccessRate> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		build,
    		mounted,
    		successRate,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('mounted' in $$props) $$invalidate(2, mounted = $$props.mounted);
    		if ('successRate' in $$props) $$invalidate(0, successRate = $$props.successRate);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 6) {
    			data && mounted && build();
    		}
    	};

    	return [successRate, data, mounted];
    }

    class SuccessRate extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$i, create_fragment$i, safe_not_equal, { data: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "SuccessRate",
    			options,
    			id: create_fragment$i.name
    		});
    	}

    	get data() {
    		throw new Error("<SuccessRate>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<SuccessRate>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\activity\ActivityRequests.svelte generated by Svelte v3.53.1 */
    const file$g = "src\\components\\dashboard\\activity\\ActivityRequests.svelte";

    function create_fragment$h(ctx) {
    	let div1;
    	let div0;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { id: true });
    			var div0_nodes = children(div0);
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "id", "plotDiv");
    			add_location(div0, file$g, 130, 2, 3363);
    			attr_dev(div1, "id", "plotly");
    			add_location(div1, file$g, 129, 0, 3342);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			/*div0_binding*/ ctx[4](div0);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			/*div0_binding*/ ctx[4](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$h.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function periodToDays$3(period) {
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

    function instance$h($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('ActivityRequests', slots, []);

    	function defaultLayout() {
    		let periodAgo = new Date();
    		let days = periodToDays$3(period);

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
    		let days = periodToDays$3(period);

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
    		$$invalidate(3, mounted = true);
    	});

    	let { data, period } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<ActivityRequests> was created without expected prop 'data'");
    		}

    		if (period === undefined && !('period' in $$props || $$self.$$.bound[$$self.$$.props['period']])) {
    			console.warn("<ActivityRequests> was created without expected prop 'period'");
    		}
    	});

    	const writable_props = ['data', 'period'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<ActivityRequests> was created with unknown prop '${key}'`);
    	});

    	function div0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(0, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		periodToDays: periodToDays$3,
    		defaultLayout,
    		bars,
    		buildPlotData,
    		plotDiv,
    		genPlot,
    		mounted,
    		data,
    		period
    	});

    	$$self.$inject_state = $$props => {
    		if ('plotDiv' in $$props) $$invalidate(0, plotDiv = $$props.plotDiv);
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 10) {
    			data && mounted && genPlot();
    		}
    	};

    	return [plotDiv, data, period, mounted, div0_binding];
    }

    class ActivityRequests extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$h, create_fragment$h, safe_not_equal, { data: 1, period: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "ActivityRequests",
    			options,
    			id: create_fragment$h.name
    		});
    	}

    	get data() {
    		throw new Error("<ActivityRequests>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<ActivityRequests>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get period() {
    		throw new Error("<ActivityRequests>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set period(value) {
    		throw new Error("<ActivityRequests>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\activity\ActivityResponseTime.svelte generated by Svelte v3.53.1 */
    const file$f = "src\\components\\dashboard\\activity\\ActivityResponseTime.svelte";

    function create_fragment$g(ctx) {
    	let div1;
    	let div0;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { id: true });
    			var div0_nodes = children(div0);
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "id", "plotDiv");
    			add_location(div0, file$f, 149, 2, 4154);
    			attr_dev(div1, "id", "plotly");
    			add_location(div1, file$f, 148, 0, 4133);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			/*div0_binding*/ ctx[4](div0);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			/*div0_binding*/ ctx[4](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$g.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function periodToDays$2(period) {
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

    function instance$g($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('ActivityResponseTime', slots, []);

    	function defaultLayout() {
    		let periodAgo = new Date();
    		let days = periodToDays$2(period);

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
    		let days = periodToDays$2(period);

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
    		$$invalidate(3, mounted = true);
    	});

    	let { data, period } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<ActivityResponseTime> was created without expected prop 'data'");
    		}

    		if (period === undefined && !('period' in $$props || $$self.$$.bound[$$self.$$.props['period']])) {
    			console.warn("<ActivityResponseTime> was created without expected prop 'period'");
    		}
    	});

    	const writable_props = ['data', 'period'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<ActivityResponseTime> was created with unknown prop '${key}'`);
    	});

    	function div0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(0, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		periodToDays: periodToDays$2,
    		defaultLayout,
    		bars,
    		buildPlotData,
    		genPlot,
    		plotDiv,
    		mounted,
    		data,
    		period
    	});

    	$$self.$inject_state = $$props => {
    		if ('plotDiv' in $$props) $$invalidate(0, plotDiv = $$props.plotDiv);
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 10) {
    			data && mounted && genPlot();
    		}
    	};

    	return [plotDiv, data, period, mounted, div0_binding];
    }

    class ActivityResponseTime extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$g, create_fragment$g, safe_not_equal, { data: 1, period: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "ActivityResponseTime",
    			options,
    			id: create_fragment$g.name
    		});
    	}

    	get data() {
    		throw new Error("<ActivityResponseTime>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<ActivityResponseTime>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get period() {
    		throw new Error("<ActivityResponseTime>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set period(value) {
    		throw new Error("<ActivityResponseTime>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\activity\ActivitySuccessRate.svelte generated by Svelte v3.53.1 */
    const file$e = "src\\components\\dashboard\\activity\\ActivitySuccessRate.svelte";

    function get_each_context$1(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[6] = list[i];
    	return child_ctx;
    }

    // (76:2) {#if successRate != undefined}
    function create_if_block$5(ctx) {
    	let div0;
    	let t0;
    	let t1;
    	let div1;
    	let each_value = /*successRate*/ ctx[0];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block$1(get_each_context$1(ctx, each_value, i));
    	}

    	const block = {
    		c: function create() {
    			div0 = element("div");
    			t0 = text("Success rate");
    			t1 = space();
    			div1 = element("div");

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			this.h();
    		},
    		l: function claim(nodes) {
    			div0 = claim_element(nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Success rate");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(nodes);
    			div1 = claim_element(nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].l(div1_nodes);
    			}

    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "success-rate-title svelte-13xx9v");
    			add_location(div0, file$e, 76, 4, 1996);
    			attr_dev(div1, "class", "errors svelte-13xx9v");
    			add_location(div1, file$e, 77, 4, 2052);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div0, anchor);
    			append_hydration_dev(div0, t0);
    			insert_hydration_dev(target, t1, anchor);
    			insert_hydration_dev(target, div1, anchor);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(div1, null);
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*Math, successRate*/ 1) {
    				each_value = /*successRate*/ ctx[0];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context$1(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block$1(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(div1, null);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div0);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(div1);
    			destroy_each(each_blocks, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$5.name,
    		type: "if",
    		source: "(76:2) {#if successRate != undefined}",
    		ctx
    	});

    	return block;
    }

    // (79:6) {#each successRate as value}
    function create_each_block$1(ctx) {
    	let div;
    	let div_class_value;
    	let div_title_value;

    	const block = {
    		c: function create() {
    			div = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true, title: true });
    			children(div).forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", div_class_value = "error level-" + (Math.floor(/*value*/ ctx[6] * 10) + 1) + " svelte-13xx9v");

    			attr_dev(div, "title", div_title_value = /*value*/ ctx[6] >= 0
    			? (/*value*/ ctx[6] * 100).toFixed(1) + "%"
    			: "No requests");

    			add_location(div, file$e, 79, 8, 2118);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*successRate*/ 1 && div_class_value !== (div_class_value = "error level-" + (Math.floor(/*value*/ ctx[6] * 10) + 1) + " svelte-13xx9v")) {
    				attr_dev(div, "class", div_class_value);
    			}

    			if (dirty & /*successRate*/ 1 && div_title_value !== (div_title_value = /*value*/ ctx[6] >= 0
    			? (/*value*/ ctx[6] * 100).toFixed(1) + "%"
    			: "No requests")) {
    				attr_dev(div, "title", div_title_value);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block$1.name,
    		type: "each",
    		source: "(79:6) {#each successRate as value}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$f(ctx) {
    	let div;
    	let if_block = /*successRate*/ ctx[0] != undefined && create_if_block$5(ctx);

    	const block = {
    		c: function create() {
    			div = element("div");
    			if (if_block) if_block.c();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true });
    			var div_nodes = children(div);
    			if (if_block) if_block.l(div_nodes);
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "success-rate-container svelte-13xx9v");
    			add_location(div, file$e, 74, 0, 1920);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    			if (if_block) if_block.m(div, null);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*successRate*/ ctx[0] != undefined) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$5(ctx);
    					if_block.c();
    					if_block.m(div, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			if (if_block) if_block.d();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$f.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function periodToDays$1(period) {
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

    function daysAgo(date) {
    	let now = new Date();
    	return Math.floor((now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000));
    }

    function instance$f($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('ActivitySuccessRate', slots, []);

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

    		$$invalidate(0, successRate = successArr);
    	}

    	function build() {
    		setSuccessRate();
    	}

    	let successRate;
    	let mounted = false;

    	onMount(() => {
    		$$invalidate(3, mounted = true);
    	});

    	let { data, period } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<ActivitySuccessRate> was created without expected prop 'data'");
    		}

    		if (period === undefined && !('period' in $$props || $$self.$$.bound[$$self.$$.props['period']])) {
    			console.warn("<ActivitySuccessRate> was created without expected prop 'period'");
    		}
    	});

    	const writable_props = ['data', 'period'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<ActivitySuccessRate> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		periodToDays: periodToDays$1,
    		daysAgo,
    		setSuccessRate,
    		build,
    		successRate,
    		mounted,
    		data,
    		period
    	});

    	$$self.$inject_state = $$props => {
    		if ('successRate' in $$props) $$invalidate(0, successRate = $$props.successRate);
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*successRate*/ 1) ;

    		if ($$self.$$.dirty & /*data, mounted*/ 10) {
    			data && mounted && build();
    		}
    	};

    	return [successRate, data, period, mounted];
    }

    class ActivitySuccessRate extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$f, create_fragment$f, safe_not_equal, { data: 1, period: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "ActivitySuccessRate",
    			options,
    			id: create_fragment$f.name
    		});
    	}

    	get data() {
    		throw new Error("<ActivitySuccessRate>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<ActivitySuccessRate>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get period() {
    		throw new Error("<ActivitySuccessRate>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set period(value) {
    		throw new Error("<ActivitySuccessRate>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\activity\Activity.svelte generated by Svelte v3.53.1 */
    const file$d = "src\\components\\dashboard\\activity\\Activity.svelte";

    function create_fragment$e(ctx) {
    	let div1;
    	let div0;
    	let t0;
    	let t1;
    	let activityrequests;
    	let t2;
    	let activityresponsetime;
    	let t3;
    	let activitysuccessrate;
    	let current;

    	activityrequests = new ActivityRequests({
    			props: {
    				data: /*data*/ ctx[0],
    				period: /*period*/ ctx[1]
    			},
    			$$inline: true
    		});

    	activityresponsetime = new ActivityResponseTime({
    			props: {
    				data: /*data*/ ctx[0],
    				period: /*period*/ ctx[1]
    			},
    			$$inline: true
    		});

    	activitysuccessrate = new ActivitySuccessRate({
    			props: {
    				data: /*data*/ ctx[0],
    				period: /*period*/ ctx[1]
    			},
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			t0 = text("Activity");
    			t1 = space();
    			create_component(activityrequests.$$.fragment);
    			t2 = space();
    			create_component(activityresponsetime.$$.fragment);
    			t3 = space();
    			create_component(activitysuccessrate.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Activity");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div1_nodes);
    			claim_component(activityrequests.$$.fragment, div1_nodes);
    			t2 = claim_space(div1_nodes);
    			claim_component(activityresponsetime.$$.fragment, div1_nodes);
    			t3 = claim_space(div1_nodes);
    			claim_component(activitysuccessrate.$$.fragment, div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$d, 7, 2, 270);
    			attr_dev(div1, "class", "card svelte-1snjwfx");
    			add_location(div1, file$d, 6, 0, 248);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div1, t1);
    			mount_component(activityrequests, div1, null);
    			append_hydration_dev(div1, t2);
    			mount_component(activityresponsetime, div1, null);
    			append_hydration_dev(div1, t3);
    			mount_component(activitysuccessrate, div1, null);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			const activityrequests_changes = {};
    			if (dirty & /*data*/ 1) activityrequests_changes.data = /*data*/ ctx[0];
    			if (dirty & /*period*/ 2) activityrequests_changes.period = /*period*/ ctx[1];
    			activityrequests.$set(activityrequests_changes);
    			const activityresponsetime_changes = {};
    			if (dirty & /*data*/ 1) activityresponsetime_changes.data = /*data*/ ctx[0];
    			if (dirty & /*period*/ 2) activityresponsetime_changes.period = /*period*/ ctx[1];
    			activityresponsetime.$set(activityresponsetime_changes);
    			const activitysuccessrate_changes = {};
    			if (dirty & /*data*/ 1) activitysuccessrate_changes.data = /*data*/ ctx[0];
    			if (dirty & /*period*/ 2) activitysuccessrate_changes.period = /*period*/ ctx[1];
    			activitysuccessrate.$set(activitysuccessrate_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(activityrequests.$$.fragment, local);
    			transition_in(activityresponsetime.$$.fragment, local);
    			transition_in(activitysuccessrate.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(activityrequests.$$.fragment, local);
    			transition_out(activityresponsetime.$$.fragment, local);
    			transition_out(activitysuccessrate.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			destroy_component(activityrequests);
    			destroy_component(activityresponsetime);
    			destroy_component(activitysuccessrate);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$e.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$e($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Activity', slots, []);
    	let { data, period } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Activity> was created without expected prop 'data'");
    		}

    		if (period === undefined && !('period' in $$props || $$self.$$.bound[$$self.$$.props['period']])) {
    			console.warn("<Activity> was created without expected prop 'period'");
    		}
    	});

    	const writable_props = ['data', 'period'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Activity> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    		if ('period' in $$props) $$invalidate(1, period = $$props.period);
    	};

    	$$self.$capture_state = () => ({
    		ActivityRequests,
    		ActivityResponseTime,
    		ActivitySuccessRate,
    		data,
    		period
    	});

    	$$self.$inject_state = $$props => {
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    		if ('period' in $$props) $$invalidate(1, period = $$props.period);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [data, period];
    }

    class Activity extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$e, create_fragment$e, safe_not_equal, { data: 0, period: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Activity",
    			options,
    			id: create_fragment$e.name
    		});
    	}

    	get data() {
    		throw new Error("<Activity>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Activity>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get period() {
    		throw new Error("<Activity>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set period(value) {
    		throw new Error("<Activity>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\Version.svelte generated by Svelte v3.53.1 */
    const file$c = "src\\components\\dashboard\\Version.svelte";

    // (101:0) {#if versions != undefined && versions.size > 1}
    function create_if_block$4(ctx) {
    	let div3;
    	let div0;
    	let t0;
    	let t1;
    	let div2;
    	let div1;

    	const block = {
    		c: function create() {
    			div3 = element("div");
    			div0 = element("div");
    			t0 = text("Version");
    			t1 = space();
    			div2 = element("div");
    			div1 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div3 = claim_element(nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			div0 = claim_element(div3_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Version");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { id: true });
    			var div2_nodes = children(div2);
    			div1 = claim_element(div2_nodes, "DIV", { id: true, class: true });
    			var div1_nodes = children(div1);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$c, 102, 4, 2499);
    			attr_dev(div1, "id", "plotDiv");
    			attr_dev(div1, "class", "svelte-jecwjn");
    			add_location(div1, file$c, 104, 6, 2567);
    			attr_dev(div2, "id", "plotly");
    			add_location(div2, file$c, 103, 4, 2542);
    			attr_dev(div3, "class", "card svelte-jecwjn");
    			add_location(div3, file$c, 101, 2, 2475);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div3, anchor);
    			append_hydration_dev(div3, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div3, t1);
    			append_hydration_dev(div3, div2);
    			append_hydration_dev(div2, div1);
    			/*div1_binding*/ ctx[4](div1);
    		},
    		p: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div3);
    			/*div1_binding*/ ctx[4](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$4.name,
    		type: "if",
    		source: "(101:0) {#if versions != undefined && versions.size > 1}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$d(ctx) {
    	let if_block_anchor;
    	let if_block = /*versions*/ ctx[0] != undefined && /*versions*/ ctx[0].size > 1 && create_if_block$4(ctx);

    	const block = {
    		c: function create() {
    			if (if_block) if_block.c();
    			if_block_anchor = empty();
    		},
    		l: function claim(nodes) {
    			if (if_block) if_block.l(nodes);
    			if_block_anchor = empty();
    		},
    		m: function mount(target, anchor) {
    			if (if_block) if_block.m(target, anchor);
    			insert_hydration_dev(target, if_block_anchor, anchor);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*versions*/ ctx[0] != undefined && /*versions*/ ctx[0].size > 1) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$4(ctx);
    					if_block.c();
    					if_block.m(if_block_anchor.parentNode, if_block_anchor);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (if_block) if_block.d(detaching);
    			if (detaching) detach_dev(if_block_anchor);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$d.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$d($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Version', slots, []);

    	function setVersions() {
    		let v = new Set();

    		for (let i = 0; i < data.length; i++) {
    			let match = data[i].path.match(/[^a-z0-9](v\d)[^a-z0-9]/i);

    			if (match) {
    				v.add(match[1]);
    			}
    		}

    		$$invalidate(0, versions = v);

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
    		$$invalidate(3, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Version> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Version> was created with unknown prop '${key}'`);
    	});

    	function div1_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(1, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(2, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		setVersions,
    		versionPlotLayout,
    		colors,
    		pieChart,
    		versionPlotData,
    		genPlot,
    		versions,
    		plotDiv,
    		mounted,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('colors' in $$props) colors = $$props.colors;
    		if ('versions' in $$props) $$invalidate(0, versions = $$props.versions);
    		if ('plotDiv' in $$props) $$invalidate(1, plotDiv = $$props.plotDiv);
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(2, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 12) {
    			data && mounted && setVersions();
    		}
    	};

    	return [versions, plotDiv, data, mounted, div1_binding];
    }

    class Version extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$d, create_fragment$d, safe_not_equal, { data: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Version",
    			options,
    			id: create_fragment$d.name
    		});
    	}

    	get data() {
    		throw new Error("<Version>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Version>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\UsageTime.svelte generated by Svelte v3.53.1 */
    const file$b = "src\\components\\dashboard\\UsageTime.svelte";

    function create_fragment$c(ctx) {
    	let div3;
    	let div0;
    	let t0;
    	let t1;
    	let div2;
    	let div1;

    	const block = {
    		c: function create() {
    			div3 = element("div");
    			div0 = element("div");
    			t0 = text("Usage time");
    			t1 = space();
    			div2 = element("div");
    			div1 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div3 = claim_element(nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			div0 = claim_element(div3_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Usage time");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { id: true });
    			var div2_nodes = children(div2);
    			div1 = claim_element(div2_nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$b, 73, 2, 2002);
    			attr_dev(div1, "id", "plotDiv");
    			add_location(div1, file$b, 75, 4, 2069);
    			attr_dev(div2, "id", "plotly");
    			add_location(div2, file$b, 74, 2, 2046);
    			attr_dev(div3, "class", "card svelte-ecbw89");
    			add_location(div3, file$b, 72, 0, 1980);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div3, anchor);
    			append_hydration_dev(div3, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div3, t1);
    			append_hydration_dev(div3, div2);
    			append_hydration_dev(div2, div1);
    			/*div1_binding*/ ctx[3](div1);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div3);
    			/*div1_binding*/ ctx[3](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$c.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$c($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('UsageTime', slots, []);

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
    		$$invalidate(2, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<UsageTime> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<UsageTime> was created with unknown prop '${key}'`);
    	});

    	function div1_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(0, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		defaultLayout: defaultLayout$1,
    		bars,
    		buildPlotData,
    		genPlot,
    		plotDiv,
    		mounted,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('plotDiv' in $$props) $$invalidate(0, plotDiv = $$props.plotDiv);
    		if ('mounted' in $$props) $$invalidate(2, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 6) {
    			data && mounted && genPlot();
    		}
    	};

    	return [plotDiv, data, mounted, div1_binding];
    }

    class UsageTime extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$c, create_fragment$c, safe_not_equal, { data: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "UsageTime",
    			options,
    			id: create_fragment$c.name
    		});
    	}

    	get data() {
    		throw new Error("<UsageTime>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<UsageTime>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\Growth.svelte generated by Svelte v3.53.1 */
    const file$a = "src\\components\\dashboard\\Growth.svelte";

    // (44:4) {#if change != undefined}
    function create_if_block$3(ctx) {
    	let div2;
    	let div0;
    	let span0;
    	let t0_value = (/*change*/ ctx[0].requests > 0 ? "+" : "") + "";
    	let t0;
    	let t1_value = /*change*/ ctx[0].requests.toFixed(1) + "";
    	let t1;
    	let t2;
    	let t3;
    	let div1;
    	let t4;
    	let t5;
    	let div5;
    	let div3;
    	let span1;
    	let t6_value = (/*change*/ ctx[0].success > 0 ? "+" : "") + "";
    	let t6;
    	let t7_value = /*change*/ ctx[0].success.toFixed(1) + "";
    	let t7;
    	let t8;
    	let t9;
    	let div4;
    	let t10;
    	let t11;
    	let div8;
    	let div6;
    	let span2;
    	let t12_value = (/*change*/ ctx[0].responseTime > 0 ? "+" : "") + "";
    	let t12;
    	let t13_value = /*change*/ ctx[0].responseTime.toFixed(1) + "";
    	let t13;
    	let t14;
    	let t15;
    	let div7;
    	let t16;

    	const block = {
    		c: function create() {
    			div2 = element("div");
    			div0 = element("div");
    			span0 = element("span");
    			t0 = text(t0_value);
    			t1 = text(t1_value);
    			t2 = text("%");
    			t3 = space();
    			div1 = element("div");
    			t4 = text("Requests");
    			t5 = space();
    			div5 = element("div");
    			div3 = element("div");
    			span1 = element("span");
    			t6 = text(t6_value);
    			t7 = text(t7_value);
    			t8 = text("%");
    			t9 = space();
    			div4 = element("div");
    			t10 = text("Success rate");
    			t11 = space();
    			div8 = element("div");
    			div6 = element("div");
    			span2 = element("span");
    			t12 = text(t12_value);
    			t13 = text(t13_value);
    			t14 = text("%");
    			t15 = space();
    			div7 = element("div");
    			t16 = text("Response time");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div2 = claim_element(nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			div0 = claim_element(div2_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			span0 = claim_element(div0_nodes, "SPAN", { style: true });
    			var span0_nodes = children(span0);
    			t0 = claim_text(span0_nodes, t0_value);
    			t1 = claim_text(span0_nodes, t1_value);
    			t2 = claim_text(span0_nodes, "%");
    			span0_nodes.forEach(detach_dev);
    			div0_nodes.forEach(detach_dev);
    			t3 = claim_space(div2_nodes);
    			div1 = claim_element(div2_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			t4 = claim_text(div1_nodes, "Requests");
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			t5 = claim_space(nodes);
    			div5 = claim_element(nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			div3 = claim_element(div5_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			span1 = claim_element(div3_nodes, "SPAN", { style: true });
    			var span1_nodes = children(span1);
    			t6 = claim_text(span1_nodes, t6_value);
    			t7 = claim_text(span1_nodes, t7_value);
    			t8 = claim_text(span1_nodes, "%");
    			span1_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			t9 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			t10 = claim_text(div4_nodes, "Success rate");
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			t11 = claim_space(nodes);
    			div8 = claim_element(nodes, "DIV", { class: true });
    			var div8_nodes = children(div8);
    			div6 = claim_element(div8_nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			span2 = claim_element(div6_nodes, "SPAN", { style: true });
    			var span2_nodes = children(span2);
    			t12 = claim_text(span2_nodes, t12_value);
    			t13 = claim_text(span2_nodes, t13_value);
    			t14 = claim_text(span2_nodes, "%");
    			span2_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			t15 = claim_space(div8_nodes);
    			div7 = claim_element(div8_nodes, "DIV", { class: true });
    			var div7_nodes = children(div7);
    			t16 = claim_text(div7_nodes, "Response time");
    			div7_nodes.forEach(detach_dev);
    			div8_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			set_style(span0, "color", /*change*/ ctx[0].requests > 0
    			? 'var(--highlight)'
    			: 'var(--red)');

    			add_location(span0, file$a, 46, 10, 1502);
    			attr_dev(div0, "class", "tile-value svelte-1v5417n");
    			add_location(div0, file$a, 45, 8, 1466);
    			attr_dev(div1, "class", "tile-label svelte-1v5417n");
    			add_location(div1, file$a, 53, 8, 1744);
    			attr_dev(div2, "class", "tile svelte-1v5417n");
    			add_location(div2, file$a, 44, 6, 1438);

    			set_style(span1, "color", /*change*/ ctx[0].success > 0
    			? 'var(--highlight)'
    			: 'var(--red)');

    			add_location(span1, file$a, 69, 10, 2254);
    			attr_dev(div3, "class", "tile-value svelte-1v5417n");
    			add_location(div3, file$a, 68, 8, 2218);
    			attr_dev(div4, "class", "tile-label svelte-1v5417n");
    			add_location(div4, file$a, 76, 8, 2493);
    			attr_dev(div5, "class", "tile svelte-1v5417n");
    			add_location(div5, file$a, 67, 6, 2190);

    			set_style(span2, "color", /*change*/ ctx[0].responseTime < 0
    			? 'var(--highlight)'
    			: 'var(--red)');

    			add_location(span2, file$a, 80, 10, 2621);
    			attr_dev(div6, "class", "tile-value svelte-1v5417n");
    			add_location(div6, file$a, 79, 8, 2585);
    			attr_dev(div7, "class", "tile-label svelte-1v5417n");
    			add_location(div7, file$a, 89, 8, 2905);
    			attr_dev(div8, "class", "tile svelte-1v5417n");
    			add_location(div8, file$a, 78, 6, 2557);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div2, anchor);
    			append_hydration_dev(div2, div0);
    			append_hydration_dev(div0, span0);
    			append_hydration_dev(span0, t0);
    			append_hydration_dev(span0, t1);
    			append_hydration_dev(span0, t2);
    			append_hydration_dev(div2, t3);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div1, t4);
    			insert_hydration_dev(target, t5, anchor);
    			insert_hydration_dev(target, div5, anchor);
    			append_hydration_dev(div5, div3);
    			append_hydration_dev(div3, span1);
    			append_hydration_dev(span1, t6);
    			append_hydration_dev(span1, t7);
    			append_hydration_dev(span1, t8);
    			append_hydration_dev(div5, t9);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, t10);
    			insert_hydration_dev(target, t11, anchor);
    			insert_hydration_dev(target, div8, anchor);
    			append_hydration_dev(div8, div6);
    			append_hydration_dev(div6, span2);
    			append_hydration_dev(span2, t12);
    			append_hydration_dev(span2, t13);
    			append_hydration_dev(span2, t14);
    			append_hydration_dev(div8, t15);
    			append_hydration_dev(div8, div7);
    			append_hydration_dev(div7, t16);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*change*/ 1 && t0_value !== (t0_value = (/*change*/ ctx[0].requests > 0 ? "+" : "") + "")) set_data_dev(t0, t0_value);
    			if (dirty & /*change*/ 1 && t1_value !== (t1_value = /*change*/ ctx[0].requests.toFixed(1) + "")) set_data_dev(t1, t1_value);

    			if (dirty & /*change*/ 1) {
    				set_style(span0, "color", /*change*/ ctx[0].requests > 0
    				? 'var(--highlight)'
    				: 'var(--red)');
    			}

    			if (dirty & /*change*/ 1 && t6_value !== (t6_value = (/*change*/ ctx[0].success > 0 ? "+" : "") + "")) set_data_dev(t6, t6_value);
    			if (dirty & /*change*/ 1 && t7_value !== (t7_value = /*change*/ ctx[0].success.toFixed(1) + "")) set_data_dev(t7, t7_value);

    			if (dirty & /*change*/ 1) {
    				set_style(span1, "color", /*change*/ ctx[0].success > 0
    				? 'var(--highlight)'
    				: 'var(--red)');
    			}

    			if (dirty & /*change*/ 1 && t12_value !== (t12_value = (/*change*/ ctx[0].responseTime > 0 ? "+" : "") + "")) set_data_dev(t12, t12_value);
    			if (dirty & /*change*/ 1 && t13_value !== (t13_value = /*change*/ ctx[0].responseTime.toFixed(1) + "")) set_data_dev(t13, t13_value);

    			if (dirty & /*change*/ 1) {
    				set_style(span2, "color", /*change*/ ctx[0].responseTime < 0
    				? 'var(--highlight)'
    				: 'var(--red)');
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div2);
    			if (detaching) detach_dev(t5);
    			if (detaching) detach_dev(div5);
    			if (detaching) detach_dev(t11);
    			if (detaching) detach_dev(div8);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$3.name,
    		type: "if",
    		source: "(44:4) {#if change != undefined}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$b(ctx) {
    	let div2;
    	let div0;
    	let t0;
    	let t1;
    	let div1;
    	let if_block = /*change*/ ctx[0] != undefined && create_if_block$3(ctx);

    	const block = {
    		c: function create() {
    			div2 = element("div");
    			div0 = element("div");
    			t0 = text("Growth");
    			t1 = space();
    			div1 = element("div");
    			if (if_block) if_block.c();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div2 = claim_element(nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			div0 = claim_element(div2_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Growth");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div2_nodes);
    			div1 = claim_element(div2_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			if (if_block) if_block.l(div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$a, 41, 2, 1339);
    			attr_dev(div1, "class", "values svelte-1v5417n");
    			add_location(div1, file$a, 42, 2, 1379);
    			attr_dev(div2, "class", "card svelte-1v5417n");
    			add_location(div2, file$a, 40, 0, 1317);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div2, anchor);
    			append_hydration_dev(div2, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div2, t1);
    			append_hydration_dev(div2, div1);
    			if (if_block) if_block.m(div1, null);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*change*/ ctx[0] != undefined) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$3(ctx);
    					if_block.c();
    					if_block.m(div1, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div2);
    			if (if_block) if_block.d();
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$b.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$b($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Growth', slots, []);

    	function buildWeek() {
    		let thisPeriod = periodData(data);
    		let lastPeriod = periodData(prevData);
    		let requestsChange = (thisPeriod.requests + 1) / (lastPeriod.requests + 1) * 100 - 100;
    		let successChange = ((thisPeriod.success + 1) / (thisPeriod.requests + 1) + 1) / ((lastPeriod.success + 1) / (lastPeriod.requests + 1) + 1) * 100 - 100;
    		let responseTimeChange = ((thisPeriod.responseTime + 1) / (thisPeriod.requests + 1) + 1) / ((lastPeriod.responseTime + 1) / (lastPeriod.requests + 1) + 1) * 100 - 100;

    		$$invalidate(0, change = {
    			requests: requestsChange,
    			success: successChange,
    			responseTime: responseTimeChange
    		});
    	}

    	let change;
    	let mounted = false;

    	onMount(() => {
    		$$invalidate(3, mounted = true);
    	});

    	let { data, prevData } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Growth> was created without expected prop 'data'");
    		}

    		if (prevData === undefined && !('prevData' in $$props || $$self.$$.bound[$$self.$$.props['prevData']])) {
    			console.warn("<Growth> was created without expected prop 'prevData'");
    		}
    	});

    	const writable_props = ['data', 'prevData'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Growth> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('prevData' in $$props) $$invalidate(2, prevData = $$props.prevData);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		periodData,
    		buildWeek,
    		change,
    		mounted,
    		data,
    		prevData
    	});

    	$$self.$inject_state = $$props => {
    		if ('change' in $$props) $$invalidate(0, change = $$props.change);
    		if ('mounted' in $$props) $$invalidate(3, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('prevData' in $$props) $$invalidate(2, prevData = $$props.prevData);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 10) {
    			data && mounted && buildWeek();
    		}
    	};

    	return [change, data, prevData, mounted];
    }

    class Growth extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$b, create_fragment$b, safe_not_equal, { data: 1, prevData: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Growth",
    			options,
    			id: create_fragment$b.name
    		});
    	}

    	get data() {
    		throw new Error("<Growth>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Growth>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get prevData() {
    		throw new Error("<Growth>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set prevData(value) {
    		throw new Error("<Growth>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\device\Browser.svelte generated by Svelte v3.53.1 */
    const file$9 = "src\\components\\dashboard\\device\\Browser.svelte";

    function create_fragment$a(ctx) {
    	let div1;
    	let div0;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { id: true, class: true });
    			var div0_nodes = children(div0);
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "id", "plotDiv");
    			attr_dev(div0, "class", "svelte-njiwdy");
    			add_location(div0, file$9, 127, 2, 3194);
    			attr_dev(div1, "id", "plotly");
    			add_location(div1, file$9, 126, 0, 3173);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			/*div0_binding*/ ctx[3](div0);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			/*div0_binding*/ ctx[3](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$a.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$a($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Browser', slots, []);

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
    		$$invalidate(2, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Browser> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Browser> was created with unknown prop '${key}'`);
    	});

    	function div0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(0, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		getBrowser,
    		browserPlotLayout,
    		colors,
    		pieChart,
    		browserPlotData,
    		genPlot,
    		plotDiv,
    		mounted,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('colors' in $$props) colors = $$props.colors;
    		if ('plotDiv' in $$props) $$invalidate(0, plotDiv = $$props.plotDiv);
    		if ('mounted' in $$props) $$invalidate(2, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 6) {
    			data && mounted && genPlot();
    		}
    	};

    	return [plotDiv, data, mounted, div0_binding];
    }

    class Browser extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$a, create_fragment$a, safe_not_equal, { data: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Browser",
    			options,
    			id: create_fragment$a.name
    		});
    	}

    	get data() {
    		throw new Error("<Browser>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Browser>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\device\OperatingSystem.svelte generated by Svelte v3.53.1 */
    const file$8 = "src\\components\\dashboard\\device\\OperatingSystem.svelte";

    function create_fragment$9(ctx) {
    	let div1;
    	let div0;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { id: true, class: true });
    			var div0_nodes = children(div0);
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "id", "plotDiv");
    			attr_dev(div0, "class", "svelte-njiwdy");
    			add_location(div0, file$8, 154, 2, 4026);
    			attr_dev(div1, "id", "plotly");
    			add_location(div1, file$8, 153, 0, 4005);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			/*div0_binding*/ ctx[3](div0);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			/*div0_binding*/ ctx[3](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$9.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$9($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('OperatingSystem', slots, []);

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
    		$$invalidate(2, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<OperatingSystem> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<OperatingSystem> was created with unknown prop '${key}'`);
    	});

    	function div0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(0, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		getOS,
    		osPlotLayout,
    		colors,
    		pieChart,
    		osPlotData,
    		genPlot,
    		plotDiv,
    		mounted,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('colors' in $$props) colors = $$props.colors;
    		if ('plotDiv' in $$props) $$invalidate(0, plotDiv = $$props.plotDiv);
    		if ('mounted' in $$props) $$invalidate(2, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 6) {
    			data && mounted && genPlot();
    		}
    	};

    	return [plotDiv, data, mounted, div0_binding];
    }

    class OperatingSystem extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$9, create_fragment$9, safe_not_equal, { data: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "OperatingSystem",
    			options,
    			id: create_fragment$9.name
    		});
    	}

    	get data() {
    		throw new Error("<OperatingSystem>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<OperatingSystem>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\device\DeviceType.svelte generated by Svelte v3.53.1 */
    const file$7 = "src\\components\\dashboard\\device\\DeviceType.svelte";

    function create_fragment$8(ctx) {
    	let div1;
    	let div0;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { id: true, class: true });
    			var div0_nodes = children(div0);
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "id", "plotDiv");
    			attr_dev(div0, "class", "svelte-njiwdy");
    			add_location(div0, file$7, 94, 2, 2230);
    			attr_dev(div1, "id", "plotly");
    			add_location(div1, file$7, 93, 0, 2209);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			/*div0_binding*/ ctx[3](div0);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			/*div0_binding*/ ctx[3](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$8.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$8($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('DeviceType', slots, []);
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
    		$$invalidate(2, mounted = true);
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<DeviceType> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<DeviceType> was created with unknown prop '${key}'`);
    	});

    	function div0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(0, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		getDevice,
    		devicePlotLayout,
    		colors,
    		pieChart,
    		devicePlotData,
    		genPlot,
    		plotDiv,
    		mounted,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('colors' in $$props) colors = $$props.colors;
    		if ('plotDiv' in $$props) $$invalidate(0, plotDiv = $$props.plotDiv);
    		if ('mounted' in $$props) $$invalidate(2, mounted = $$props.mounted);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*data, mounted*/ 6) {
    			data && mounted && genPlot();
    		}
    	};

    	return [plotDiv, data, mounted, div0_binding];
    }

    class DeviceType extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$8, create_fragment$8, safe_not_equal, { data: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "DeviceType",
    			options,
    			id: create_fragment$8.name
    		});
    	}

    	get data() {
    		throw new Error("<DeviceType>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<DeviceType>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\dashboard\device\Device.svelte generated by Svelte v3.53.1 */
    const file$6 = "src\\components\\dashboard\\device\\Device.svelte";

    function create_fragment$7(ctx) {
    	let div5;
    	let div1;
    	let t0;
    	let div0;
    	let button0;
    	let t1;
    	let button0_class_value;
    	let t2;
    	let button1;
    	let t3;
    	let button1_class_value;
    	let t4;
    	let button2;
    	let t5;
    	let button2_class_value;
    	let t6;
    	let div2;
    	let operatingsystem;
    	let div2_style_value;
    	let t7;
    	let div3;
    	let browser;
    	let div3_style_value;
    	let t8;
    	let div4;
    	let devicetype;
    	let div4_style_value;
    	let current;
    	let mounted;
    	let dispose;

    	operatingsystem = new OperatingSystem({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	browser = new Browser({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	devicetype = new DeviceType({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			div5 = element("div");
    			div1 = element("div");
    			t0 = text("Device\r\n\r\n    ");
    			div0 = element("div");
    			button0 = element("button");
    			t1 = text("OS");
    			t2 = space();
    			button1 = element("button");
    			t3 = text("Browser");
    			t4 = space();
    			button2 = element("button");
    			t5 = text("Device");
    			t6 = space();
    			div2 = element("div");
    			create_component(operatingsystem.$$.fragment);
    			t7 = space();
    			div3 = element("div");
    			create_component(browser.$$.fragment);
    			t8 = space();
    			div4 = element("div");
    			create_component(devicetype.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div5 = claim_element(nodes, "DIV", { class: true, title: true });
    			var div5_nodes = children(div5);
    			div1 = claim_element(div5_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			t0 = claim_text(div1_nodes, "Device\r\n\r\n    ");
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			button0 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button0_nodes = children(button0);
    			t1 = claim_text(button0_nodes, "OS");
    			button0_nodes.forEach(detach_dev);
    			t2 = claim_space(div0_nodes);
    			button1 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button1_nodes = children(button1);
    			t3 = claim_text(button1_nodes, "Browser");
    			button1_nodes.forEach(detach_dev);
    			t4 = claim_space(div0_nodes);
    			button2 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button2_nodes = children(button2);
    			t5 = claim_text(button2_nodes, "Device");
    			button2_nodes.forEach(detach_dev);
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			t6 = claim_space(div5_nodes);
    			div2 = claim_element(div5_nodes, "DIV", { class: true, style: true });
    			var div2_nodes = children(div2);
    			claim_component(operatingsystem.$$.fragment, div2_nodes);
    			div2_nodes.forEach(detach_dev);
    			t7 = claim_space(div5_nodes);
    			div3 = claim_element(div5_nodes, "DIV", { class: true, style: true });
    			var div3_nodes = children(div3);
    			claim_component(browser.$$.fragment, div3_nodes);
    			div3_nodes.forEach(detach_dev);
    			t8 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true, style: true });
    			var div4_nodes = children(div4);
    			claim_component(devicetype.$$.fragment, div4_nodes);
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(button0, "class", button0_class_value = "" + (null_to_empty(/*activeBtn*/ ctx[1] == "os" ? "active" : "") + " svelte-sfsn3p"));
    			add_location(button0, file$6, 17, 6, 508);
    			attr_dev(button1, "class", button1_class_value = "" + (null_to_empty(/*activeBtn*/ ctx[1] == "browser" ? "active" : "") + " svelte-sfsn3p"));
    			add_location(button1, file$6, 23, 6, 658);
    			attr_dev(button2, "class", button2_class_value = "" + (null_to_empty(/*activeBtn*/ ctx[1] == "device" ? "active" : "") + " svelte-sfsn3p"));
    			add_location(button2, file$6, 29, 6, 823);
    			attr_dev(div0, "class", "toggle svelte-sfsn3p");
    			add_location(div0, file$6, 16, 4, 480);
    			attr_dev(div1, "class", "card-title svelte-sfsn3p");
    			add_location(div1, file$6, 13, 2, 436);
    			attr_dev(div2, "class", "os svelte-sfsn3p");
    			attr_dev(div2, "style", div2_style_value = /*activeBtn*/ ctx[1] == "os" ? "display: initial" : "");
    			add_location(div2, file$6, 37, 2, 1003);
    			attr_dev(div3, "class", "browser svelte-sfsn3p");

    			attr_dev(div3, "style", div3_style_value = /*activeBtn*/ ctx[1] == "browser"
    			? "display: initial"
    			: "");

    			add_location(div3, file$6, 40, 2, 1117);
    			attr_dev(div4, "class", "device svelte-sfsn3p");

    			attr_dev(div4, "style", div4_style_value = /*activeBtn*/ ctx[1] == "device"
    			? "display: initial"
    			: "");

    			add_location(div4, file$6, 43, 2, 1233);
    			attr_dev(div5, "class", "card svelte-sfsn3p");
    			attr_dev(div5, "title", "Last week");
    			add_location(div5, file$6, 12, 0, 396);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div5, anchor);
    			append_hydration_dev(div5, div1);
    			append_hydration_dev(div1, t0);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, button0);
    			append_hydration_dev(button0, t1);
    			append_hydration_dev(div0, t2);
    			append_hydration_dev(div0, button1);
    			append_hydration_dev(button1, t3);
    			append_hydration_dev(div0, t4);
    			append_hydration_dev(div0, button2);
    			append_hydration_dev(button2, t5);
    			append_hydration_dev(div5, t6);
    			append_hydration_dev(div5, div2);
    			mount_component(operatingsystem, div2, null);
    			append_hydration_dev(div5, t7);
    			append_hydration_dev(div5, div3);
    			mount_component(browser, div3, null);
    			append_hydration_dev(div5, t8);
    			append_hydration_dev(div5, div4);
    			mount_component(devicetype, div4, null);
    			current = true;

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*click_handler*/ ctx[3], false, false, false),
    					listen_dev(button1, "click", /*click_handler_1*/ ctx[4], false, false, false),
    					listen_dev(button2, "click", /*click_handler_2*/ ctx[5], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (!current || dirty & /*activeBtn*/ 2 && button0_class_value !== (button0_class_value = "" + (null_to_empty(/*activeBtn*/ ctx[1] == "os" ? "active" : "") + " svelte-sfsn3p"))) {
    				attr_dev(button0, "class", button0_class_value);
    			}

    			if (!current || dirty & /*activeBtn*/ 2 && button1_class_value !== (button1_class_value = "" + (null_to_empty(/*activeBtn*/ ctx[1] == "browser" ? "active" : "") + " svelte-sfsn3p"))) {
    				attr_dev(button1, "class", button1_class_value);
    			}

    			if (!current || dirty & /*activeBtn*/ 2 && button2_class_value !== (button2_class_value = "" + (null_to_empty(/*activeBtn*/ ctx[1] == "device" ? "active" : "") + " svelte-sfsn3p"))) {
    				attr_dev(button2, "class", button2_class_value);
    			}

    			const operatingsystem_changes = {};
    			if (dirty & /*data*/ 1) operatingsystem_changes.data = /*data*/ ctx[0];
    			operatingsystem.$set(operatingsystem_changes);

    			if (!current || dirty & /*activeBtn*/ 2 && div2_style_value !== (div2_style_value = /*activeBtn*/ ctx[1] == "os" ? "display: initial" : "")) {
    				attr_dev(div2, "style", div2_style_value);
    			}

    			const browser_changes = {};
    			if (dirty & /*data*/ 1) browser_changes.data = /*data*/ ctx[0];
    			browser.$set(browser_changes);

    			if (!current || dirty & /*activeBtn*/ 2 && div3_style_value !== (div3_style_value = /*activeBtn*/ ctx[1] == "browser"
    			? "display: initial"
    			: "")) {
    				attr_dev(div3, "style", div3_style_value);
    			}

    			const devicetype_changes = {};
    			if (dirty & /*data*/ 1) devicetype_changes.data = /*data*/ ctx[0];
    			devicetype.$set(devicetype_changes);

    			if (!current || dirty & /*activeBtn*/ 2 && div4_style_value !== (div4_style_value = /*activeBtn*/ ctx[1] == "device"
    			? "display: initial"
    			: "")) {
    				attr_dev(div4, "style", div4_style_value);
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(operatingsystem.$$.fragment, local);
    			transition_in(browser.$$.fragment, local);
    			transition_in(devicetype.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(operatingsystem.$$.fragment, local);
    			transition_out(browser.$$.fragment, local);
    			transition_out(devicetype.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div5);
    			destroy_component(operatingsystem);
    			destroy_component(browser);
    			destroy_component(devicetype);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$7.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$7($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Device', slots, []);

    	function setBtn(target) {
    		$$invalidate(1, activeBtn = target);

    		// Resize window to trigger new plot resize to match current card size
    		window.dispatchEvent(new Event("resize"));
    	}

    	let activeBtn = "os";
    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Device> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Device> was created with unknown prop '${key}'`);
    	});

    	const click_handler = () => {
    		setBtn("os");
    	};

    	const click_handler_1 = () => {
    		setBtn("browser");
    	};

    	const click_handler_2 = () => {
    		setBtn("device");
    	};

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		Browser,
    		OperatingSystem,
    		DeviceType,
    		setBtn,
    		activeBtn,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('activeBtn' in $$props) $$invalidate(1, activeBtn = $$props.activeBtn);
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [data, activeBtn, setBtn, click_handler, click_handler_1, click_handler_2];
    }

    class Device extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$7, create_fragment$7, safe_not_equal, { data: 0 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Device",
    			options,
    			id: create_fragment$7.name
    		});
    	}

    	get data() {
    		throw new Error("<Device>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Device>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\routes\Dashboard.svelte generated by Svelte v3.53.1 */

    const { console: console_1$1 } = globals;
    const file$5 = "src\\routes\\Dashboard.svelte";

    // (342:0) {:else}
    function create_else_block$2(ctx) {
    	let div2;
    	let div1;
    	let div0;

    	const block = {
    		c: function create() {
    			div2 = element("div");
    			div1 = element("div");
    			div0 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div2 = claim_element(nodes, "DIV", { class: true, style: true });
    			var div2_nodes = children(div2);
    			div1 = claim_element(div2_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			children(div0).forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "loader");
    			add_location(div0, file$5, 344, 6, 11465);
    			attr_dev(div1, "class", "spinner");
    			add_location(div1, file$5, 343, 4, 11436);
    			attr_dev(div2, "class", "placeholder svelte-1j1sa1i");
    			set_style(div2, "min-height", "85vh");
    			add_location(div2, file$5, 342, 2, 11379);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div2, anchor);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div1, div0);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div2);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_else_block$2.name,
    		type: "else",
    		source: "(342:0) {:else}",
    		ctx
    	});

    	return block;
    }

    // (340:17) 
    function create_if_block_1$1(ctx) {
    	let div;
    	let t;

    	const block = {
    		c: function create() {
    			div = element("div");
    			t = text("No requests currently logged.");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true });
    			var div_nodes = children(div);
    			t = claim_text(div_nodes, "No requests currently logged.");
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "no-requests svelte-1j1sa1i");
    			add_location(div, file$5, 340, 2, 11306);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    			append_hydration_dev(div, t);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1$1.name,
    		type: "if",
    		source: "(340:17) ",
    		ctx
    	});

    	return block;
    }

    // (244:0) {#if periodData != undefined}
    function create_if_block$2(ctx) {
    	let div6;
    	let div0;
    	let button0;
    	let t0;
    	let button0_class_value;
    	let t1;
    	let button1;
    	let t2;
    	let button1_class_value;
    	let t3;
    	let button2;
    	let t4;
    	let button2_class_value;
    	let t5;
    	let button3;
    	let t6;
    	let button3_class_value;
    	let t7;
    	let button4;
    	let t8;
    	let button4_class_value;
    	let t9;
    	let button5;
    	let t10;
    	let button5_class_value;
    	let t11;
    	let button6;
    	let t12;
    	let button6_class_value;
    	let t13;
    	let div3;
    	let div1;
    	let welcome;
    	let t14;
    	let successrate;
    	let t15;
    	let div2;
    	let requests;
    	let t16;
    	let requestsperhour;
    	let t17;
    	let responsetimes;
    	let t18;
    	let endpoints;
    	let t19;
    	let version;
    	let t20;
    	let div5;
    	let activity;
    	let t21;
    	let div4;
    	let growth;
    	let t22;
    	let device;
    	let t23;
    	let usagetime;
    	let current;
    	let mounted;
    	let dispose;
    	welcome = new Welcome({ $$inline: true });

    	successrate = new SuccessRate({
    			props: { data: /*periodData*/ ctx[0] },
    			$$inline: true
    		});

    	requests = new Requests({
    			props: {
    				data: /*periodData*/ ctx[0],
    				prevData: /*prevPeriodData*/ ctx[1]
    			},
    			$$inline: true
    		});

    	requestsperhour = new RequestsPerHour({
    			props: {
    				data: /*periodData*/ ctx[0],
    				period: /*period*/ ctx[2]
    			},
    			$$inline: true
    		});

    	responsetimes = new ResponseTimes({
    			props: { data: /*periodData*/ ctx[0] },
    			$$inline: true
    		});

    	endpoints = new Endpoints({
    			props: { data: /*periodData*/ ctx[0] },
    			$$inline: true
    		});

    	version = new Version({
    			props: { data: /*periodData*/ ctx[0] },
    			$$inline: true
    		});

    	activity = new Activity({
    			props: {
    				data: /*periodData*/ ctx[0],
    				period: /*period*/ ctx[2]
    			},
    			$$inline: true
    		});

    	growth = new Growth({
    			props: {
    				data: /*periodData*/ ctx[0],
    				prevData: /*prevPeriodData*/ ctx[1]
    			},
    			$$inline: true
    		});

    	device = new Device({
    			props: { data: /*periodData*/ ctx[0] },
    			$$inline: true
    		});

    	usagetime = new UsageTime({
    			props: { data: /*periodData*/ ctx[0] },
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			div6 = element("div");
    			div0 = element("div");
    			button0 = element("button");
    			t0 = text("24 hours");
    			t1 = space();
    			button1 = element("button");
    			t2 = text("Week");
    			t3 = space();
    			button2 = element("button");
    			t4 = text("Month");
    			t5 = space();
    			button3 = element("button");
    			t6 = text("3 months");
    			t7 = space();
    			button4 = element("button");
    			t8 = text("6 months");
    			t9 = space();
    			button5 = element("button");
    			t10 = text("Year");
    			t11 = space();
    			button6 = element("button");
    			t12 = text("All time");
    			t13 = space();
    			div3 = element("div");
    			div1 = element("div");
    			create_component(welcome.$$.fragment);
    			t14 = space();
    			create_component(successrate.$$.fragment);
    			t15 = space();
    			div2 = element("div");
    			create_component(requests.$$.fragment);
    			t16 = space();
    			create_component(requestsperhour.$$.fragment);
    			t17 = space();
    			create_component(responsetimes.$$.fragment);
    			t18 = space();
    			create_component(endpoints.$$.fragment);
    			t19 = space();
    			create_component(version.$$.fragment);
    			t20 = space();
    			div5 = element("div");
    			create_component(activity.$$.fragment);
    			t21 = space();
    			div4 = element("div");
    			create_component(growth.$$.fragment);
    			t22 = space();
    			create_component(device.$$.fragment);
    			t23 = space();
    			create_component(usagetime.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div6 = claim_element(nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div0 = claim_element(div6_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			button0 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button0_nodes = children(button0);
    			t0 = claim_text(button0_nodes, "24 hours");
    			button0_nodes.forEach(detach_dev);
    			t1 = claim_space(div0_nodes);
    			button1 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button1_nodes = children(button1);
    			t2 = claim_text(button1_nodes, "Week");
    			button1_nodes.forEach(detach_dev);
    			t3 = claim_space(div0_nodes);
    			button2 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button2_nodes = children(button2);
    			t4 = claim_text(button2_nodes, "Month");
    			button2_nodes.forEach(detach_dev);
    			t5 = claim_space(div0_nodes);
    			button3 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button3_nodes = children(button3);
    			t6 = claim_text(button3_nodes, "3 months");
    			button3_nodes.forEach(detach_dev);
    			t7 = claim_space(div0_nodes);
    			button4 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button4_nodes = children(button4);
    			t8 = claim_text(button4_nodes, "6 months");
    			button4_nodes.forEach(detach_dev);
    			t9 = claim_space(div0_nodes);
    			button5 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button5_nodes = children(button5);
    			t10 = claim_text(button5_nodes, "Year");
    			button5_nodes.forEach(detach_dev);
    			t11 = claim_space(div0_nodes);
    			button6 = claim_element(div0_nodes, "BUTTON", { class: true });
    			var button6_nodes = children(button6);
    			t12 = claim_text(button6_nodes, "All time");
    			button6_nodes.forEach(detach_dev);
    			div0_nodes.forEach(detach_dev);
    			t13 = claim_space(div6_nodes);
    			div3 = claim_element(div6_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			div1 = claim_element(div3_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			claim_component(welcome.$$.fragment, div1_nodes);
    			t14 = claim_space(div1_nodes);
    			claim_component(successrate.$$.fragment, div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			t15 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			claim_component(requests.$$.fragment, div2_nodes);
    			t16 = claim_space(div2_nodes);
    			claim_component(requestsperhour.$$.fragment, div2_nodes);
    			div2_nodes.forEach(detach_dev);
    			t17 = claim_space(div3_nodes);
    			claim_component(responsetimes.$$.fragment, div3_nodes);
    			t18 = claim_space(div3_nodes);
    			claim_component(endpoints.$$.fragment, div3_nodes);
    			t19 = claim_space(div3_nodes);
    			claim_component(version.$$.fragment, div3_nodes);
    			div3_nodes.forEach(detach_dev);
    			t20 = claim_space(div6_nodes);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			claim_component(activity.$$.fragment, div5_nodes);
    			t21 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			claim_component(growth.$$.fragment, div4_nodes);
    			t22 = claim_space(div4_nodes);
    			claim_component(device.$$.fragment, div4_nodes);
    			div4_nodes.forEach(detach_dev);
    			t23 = claim_space(div5_nodes);
    			claim_component(usagetime.$$.fragment, div5_nodes);
    			div5_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(button0, "class", button0_class_value = "time-period-btn " + (/*period*/ ctx[2] == '24-hours'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i");

    			add_location(button0, file$5, 246, 6, 8920);

    			attr_dev(button1, "class", button1_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'week'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i");

    			add_location(button1, file$5, 256, 6, 9162);

    			attr_dev(button2, "class", button2_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'month'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i");

    			add_location(button2, file$5, 266, 6, 9392);

    			attr_dev(button3, "class", button3_class_value = "time-period-btn " + (/*period*/ ctx[2] == '3-months'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i");

    			add_location(button3, file$5, 276, 6, 9625);

    			attr_dev(button4, "class", button4_class_value = "time-period-btn " + (/*period*/ ctx[2] == '6-months'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i");

    			add_location(button4, file$5, 286, 6, 9867);

    			attr_dev(button5, "class", button5_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'year'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i");

    			add_location(button5, file$5, 296, 6, 10109);

    			attr_dev(button6, "class", button6_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'all-time'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i");

    			add_location(button6, file$5, 306, 6, 10339);
    			attr_dev(div0, "class", "time-period svelte-1j1sa1i");
    			add_location(div0, file$5, 245, 4, 8887);
    			attr_dev(div1, "class", "row svelte-1j1sa1i");
    			add_location(div1, file$5, 318, 6, 10617);
    			attr_dev(div2, "class", "row svelte-1j1sa1i");
    			add_location(div2, file$5, 322, 6, 10720);
    			attr_dev(div3, "class", "left");
    			add_location(div3, file$5, 317, 4, 10591);
    			attr_dev(div4, "class", "grid-row svelte-1j1sa1i");
    			add_location(div4, file$5, 332, 6, 11084);
    			attr_dev(div5, "class", "right svelte-1j1sa1i");
    			add_location(div5, file$5, 330, 4, 11010);
    			attr_dev(div6, "class", "dashboard svelte-1j1sa1i");
    			add_location(div6, file$5, 244, 2, 8858);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div6, anchor);
    			append_hydration_dev(div6, div0);
    			append_hydration_dev(div0, button0);
    			append_hydration_dev(button0, t0);
    			append_hydration_dev(div0, t1);
    			append_hydration_dev(div0, button1);
    			append_hydration_dev(button1, t2);
    			append_hydration_dev(div0, t3);
    			append_hydration_dev(div0, button2);
    			append_hydration_dev(button2, t4);
    			append_hydration_dev(div0, t5);
    			append_hydration_dev(div0, button3);
    			append_hydration_dev(button3, t6);
    			append_hydration_dev(div0, t7);
    			append_hydration_dev(div0, button4);
    			append_hydration_dev(button4, t8);
    			append_hydration_dev(div0, t9);
    			append_hydration_dev(div0, button5);
    			append_hydration_dev(button5, t10);
    			append_hydration_dev(div0, t11);
    			append_hydration_dev(div0, button6);
    			append_hydration_dev(button6, t12);
    			append_hydration_dev(div6, t13);
    			append_hydration_dev(div6, div3);
    			append_hydration_dev(div3, div1);
    			mount_component(welcome, div1, null);
    			append_hydration_dev(div1, t14);
    			mount_component(successrate, div1, null);
    			append_hydration_dev(div3, t15);
    			append_hydration_dev(div3, div2);
    			mount_component(requests, div2, null);
    			append_hydration_dev(div2, t16);
    			mount_component(requestsperhour, div2, null);
    			append_hydration_dev(div3, t17);
    			mount_component(responsetimes, div3, null);
    			append_hydration_dev(div3, t18);
    			mount_component(endpoints, div3, null);
    			append_hydration_dev(div3, t19);
    			mount_component(version, div3, null);
    			append_hydration_dev(div6, t20);
    			append_hydration_dev(div6, div5);
    			mount_component(activity, div5, null);
    			append_hydration_dev(div5, t21);
    			append_hydration_dev(div5, div4);
    			mount_component(growth, div4, null);
    			append_hydration_dev(div4, t22);
    			mount_component(device, div4, null);
    			append_hydration_dev(div5, t23);
    			mount_component(usagetime, div5, null);
    			current = true;

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*click_handler*/ ctx[7], false, false, false),
    					listen_dev(button1, "click", /*click_handler_1*/ ctx[8], false, false, false),
    					listen_dev(button2, "click", /*click_handler_2*/ ctx[9], false, false, false),
    					listen_dev(button3, "click", /*click_handler_3*/ ctx[10], false, false, false),
    					listen_dev(button4, "click", /*click_handler_4*/ ctx[11], false, false, false),
    					listen_dev(button5, "click", /*click_handler_5*/ ctx[12], false, false, false),
    					listen_dev(button6, "click", /*click_handler_6*/ ctx[13], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, dirty) {
    			if (!current || dirty & /*period*/ 4 && button0_class_value !== (button0_class_value = "time-period-btn " + (/*period*/ ctx[2] == '24-hours'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i")) {
    				attr_dev(button0, "class", button0_class_value);
    			}

    			if (!current || dirty & /*period*/ 4 && button1_class_value !== (button1_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'week'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i")) {
    				attr_dev(button1, "class", button1_class_value);
    			}

    			if (!current || dirty & /*period*/ 4 && button2_class_value !== (button2_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'month'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i")) {
    				attr_dev(button2, "class", button2_class_value);
    			}

    			if (!current || dirty & /*period*/ 4 && button3_class_value !== (button3_class_value = "time-period-btn " + (/*period*/ ctx[2] == '3-months'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i")) {
    				attr_dev(button3, "class", button3_class_value);
    			}

    			if (!current || dirty & /*period*/ 4 && button4_class_value !== (button4_class_value = "time-period-btn " + (/*period*/ ctx[2] == '6-months'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i")) {
    				attr_dev(button4, "class", button4_class_value);
    			}

    			if (!current || dirty & /*period*/ 4 && button5_class_value !== (button5_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'year'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i")) {
    				attr_dev(button5, "class", button5_class_value);
    			}

    			if (!current || dirty & /*period*/ 4 && button6_class_value !== (button6_class_value = "time-period-btn " + (/*period*/ ctx[2] == 'all-time'
    			? 'time-period-btn-active'
    			: '') + " svelte-1j1sa1i")) {
    				attr_dev(button6, "class", button6_class_value);
    			}

    			const successrate_changes = {};
    			if (dirty & /*periodData*/ 1) successrate_changes.data = /*periodData*/ ctx[0];
    			successrate.$set(successrate_changes);
    			const requests_changes = {};
    			if (dirty & /*periodData*/ 1) requests_changes.data = /*periodData*/ ctx[0];
    			if (dirty & /*prevPeriodData*/ 2) requests_changes.prevData = /*prevPeriodData*/ ctx[1];
    			requests.$set(requests_changes);
    			const requestsperhour_changes = {};
    			if (dirty & /*periodData*/ 1) requestsperhour_changes.data = /*periodData*/ ctx[0];
    			if (dirty & /*period*/ 4) requestsperhour_changes.period = /*period*/ ctx[2];
    			requestsperhour.$set(requestsperhour_changes);
    			const responsetimes_changes = {};
    			if (dirty & /*periodData*/ 1) responsetimes_changes.data = /*periodData*/ ctx[0];
    			responsetimes.$set(responsetimes_changes);
    			const endpoints_changes = {};
    			if (dirty & /*periodData*/ 1) endpoints_changes.data = /*periodData*/ ctx[0];
    			endpoints.$set(endpoints_changes);
    			const version_changes = {};
    			if (dirty & /*periodData*/ 1) version_changes.data = /*periodData*/ ctx[0];
    			version.$set(version_changes);
    			const activity_changes = {};
    			if (dirty & /*periodData*/ 1) activity_changes.data = /*periodData*/ ctx[0];
    			if (dirty & /*period*/ 4) activity_changes.period = /*period*/ ctx[2];
    			activity.$set(activity_changes);
    			const growth_changes = {};
    			if (dirty & /*periodData*/ 1) growth_changes.data = /*periodData*/ ctx[0];
    			if (dirty & /*prevPeriodData*/ 2) growth_changes.prevData = /*prevPeriodData*/ ctx[1];
    			growth.$set(growth_changes);
    			const device_changes = {};
    			if (dirty & /*periodData*/ 1) device_changes.data = /*periodData*/ ctx[0];
    			device.$set(device_changes);
    			const usagetime_changes = {};
    			if (dirty & /*periodData*/ 1) usagetime_changes.data = /*periodData*/ ctx[0];
    			usagetime.$set(usagetime_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(welcome.$$.fragment, local);
    			transition_in(successrate.$$.fragment, local);
    			transition_in(requests.$$.fragment, local);
    			transition_in(requestsperhour.$$.fragment, local);
    			transition_in(responsetimes.$$.fragment, local);
    			transition_in(endpoints.$$.fragment, local);
    			transition_in(version.$$.fragment, local);
    			transition_in(activity.$$.fragment, local);
    			transition_in(growth.$$.fragment, local);
    			transition_in(device.$$.fragment, local);
    			transition_in(usagetime.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(welcome.$$.fragment, local);
    			transition_out(successrate.$$.fragment, local);
    			transition_out(requests.$$.fragment, local);
    			transition_out(requestsperhour.$$.fragment, local);
    			transition_out(responsetimes.$$.fragment, local);
    			transition_out(endpoints.$$.fragment, local);
    			transition_out(version.$$.fragment, local);
    			transition_out(activity.$$.fragment, local);
    			transition_out(growth.$$.fragment, local);
    			transition_out(device.$$.fragment, local);
    			transition_out(usagetime.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div6);
    			destroy_component(welcome);
    			destroy_component(successrate);
    			destroy_component(requests);
    			destroy_component(requestsperhour);
    			destroy_component(responsetimes);
    			destroy_component(endpoints);
    			destroy_component(version);
    			destroy_component(activity);
    			destroy_component(growth);
    			destroy_component(device);
    			destroy_component(usagetime);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$2.name,
    		type: "if",
    		source: "(244:0) {#if periodData != undefined}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$6(ctx) {
    	let current_block_type_index;
    	let if_block;
    	let t;
    	let footer;
    	let current;
    	const if_block_creators = [create_if_block$2, create_if_block_1$1, create_else_block$2];
    	const if_blocks = [];

    	function select_block_type(ctx, dirty) {
    		if (/*periodData*/ ctx[0] != undefined) return 0;
    		if (/*failed*/ ctx[3]) return 1;
    		return 2;
    	}

    	current_block_type_index = select_block_type(ctx);
    	if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);
    	footer = new Footer({ $$inline: true });

    	const block = {
    		c: function create() {
    			if_block.c();
    			t = space();
    			create_component(footer.$$.fragment);
    		},
    		l: function claim(nodes) {
    			if_block.l(nodes);
    			t = claim_space(nodes);
    			claim_component(footer.$$.fragment, nodes);
    		},
    		m: function mount(target, anchor) {
    			if_blocks[current_block_type_index].m(target, anchor);
    			insert_hydration_dev(target, t, anchor);
    			mount_component(footer, target, anchor);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			let previous_block_index = current_block_type_index;
    			current_block_type_index = select_block_type(ctx);

    			if (current_block_type_index === previous_block_index) {
    				if_blocks[current_block_type_index].p(ctx, dirty);
    			} else {
    				group_outros();

    				transition_out(if_blocks[previous_block_index], 1, 1, () => {
    					if_blocks[previous_block_index] = null;
    				});

    				check_outros();
    				if_block = if_blocks[current_block_type_index];

    				if (!if_block) {
    					if_block = if_blocks[current_block_type_index] = if_block_creators[current_block_type_index](ctx);
    					if_block.c();
    				} else {
    					if_block.p(ctx, dirty);
    				}

    				transition_in(if_block, 1);
    				if_block.m(t.parentNode, t);
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(if_block);
    			transition_in(footer.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block);
    			transition_out(footer.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if_blocks[current_block_type_index].d(detaching);
    			if (detaching) detach_dev(t);
    			destroy_component(footer, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$6.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$6($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Dashboard', slots, []);

    	async function fetchData() {
    		$$invalidate(5, userID = formatUUID$1(userID));

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
    			$$invalidate(3, failed = true);
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
    		$$invalidate(0, periodData = dataSubset);
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

    		$$invalidate(1, prevPeriodData = dataSubset);
    	}

    	function setPeriod(value) {
    		$$invalidate(2, period = value);
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

    	$$self.$$.on_mount.push(function () {
    		if (userID === undefined && !('userID' in $$props || $$self.$$.bound[$$self.$$.props['userID']])) {
    			console_1$1.warn("<Dashboard> was created without expected prop 'userID'");
    		}

    		if (demo === undefined && !('demo' in $$props || $$self.$$.bound[$$self.$$.props['demo']])) {
    			console_1$1.warn("<Dashboard> was created without expected prop 'demo'");
    		}
    	});

    	const writable_props = ['userID', 'demo'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console_1$1.warn(`<Dashboard> was created with unknown prop '${key}'`);
    	});

    	const click_handler = () => {
    		setPeriod("24-hours");
    	};

    	const click_handler_1 = () => {
    		setPeriod("week");
    	};

    	const click_handler_2 = () => {
    		setPeriod("month");
    	};

    	const click_handler_3 = () => {
    		setPeriod("3-months");
    	};

    	const click_handler_4 = () => {
    		setPeriod("6-months");
    	};

    	const click_handler_5 = () => {
    		setPeriod("year");
    	};

    	const click_handler_6 = () => {
    		setPeriod("all-time");
    	};

    	$$self.$$set = $$props => {
    		if ('userID' in $$props) $$invalidate(5, userID = $$props.userID);
    		if ('demo' in $$props) $$invalidate(6, demo = $$props.demo);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		Requests,
    		Welcome,
    		RequestsPerHour,
    		ResponseTimes,
    		Endpoints,
    		Footer,
    		SuccessRate,
    		Activity,
    		Version,
    		UsageTime,
    		Growth,
    		Device,
    		formatUUID: formatUUID$1,
    		fetchData,
    		inPeriod,
    		allTimePeriod,
    		periodToDays,
    		setPeriodData,
    		inPrevPeriod,
    		setPrevPeriodData,
    		setPeriod,
    		getDemoStatus,
    		getDemoUserAgent,
    		randomChoice,
    		randomChoices,
    		getHour,
    		addDemoSamples,
    		genDemoData,
    		data,
    		periodData,
    		prevPeriodData,
    		period,
    		failed,
    		userID,
    		demo
    	});

    	$$self.$inject_state = $$props => {
    		if ('data' in $$props) data = $$props.data;
    		if ('periodData' in $$props) $$invalidate(0, periodData = $$props.periodData);
    		if ('prevPeriodData' in $$props) $$invalidate(1, prevPeriodData = $$props.prevPeriodData);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    		if ('failed' in $$props) $$invalidate(3, failed = $$props.failed);
    		if ('userID' in $$props) $$invalidate(5, userID = $$props.userID);
    		if ('demo' in $$props) $$invalidate(6, demo = $$props.demo);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		periodData,
    		prevPeriodData,
    		period,
    		failed,
    		setPeriod,
    		userID,
    		demo,
    		click_handler,
    		click_handler_1,
    		click_handler_2,
    		click_handler_3,
    		click_handler_4,
    		click_handler_5,
    		click_handler_6
    	];
    }

    class Dashboard extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$6, create_fragment$6, safe_not_equal, { userID: 5, demo: 6 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Dashboard",
    			options,
    			id: create_fragment$6.name
    		});
    	}

    	get userID() {
    		throw new Error("<Dashboard>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set userID(value) {
    		throw new Error("<Dashboard>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get demo() {
    		throw new Error("<Dashboard>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set demo(value) {
    		throw new Error("<Dashboard>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\routes\Delete.svelte generated by Svelte v3.53.1 */

    const file$4 = "src\\routes\\Delete.svelte";

    function create_fragment$5(ctx) {
    	let div6;
    	let div3;
    	let h2;
    	let t0;
    	let t1;
    	let input;
    	let t2;
    	let button;
    	let t3;
    	let t4;
    	let div0;
    	let t5;
    	let t6;
    	let div2;
    	let div1;
    	let t7;
    	let div5;
    	let div4;
    	let t8;
    	let t9;
    	let img;
    	let img_src_value;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div6 = element("div");
    			div3 = element("div");
    			h2 = element("h2");
    			t0 = text("Delete all stored data");
    			t1 = space();
    			input = element("input");
    			t2 = space();
    			button = element("button");
    			t3 = text("Delete");
    			t4 = space();
    			div0 = element("div");
    			t5 = text(/*message*/ ctx[2]);
    			t6 = space();
    			div2 = element("div");
    			div1 = element("div");
    			t7 = space();
    			div5 = element("div");
    			div4 = element("div");
    			t8 = text("API Analytics");
    			t9 = space();
    			img = element("img");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div6 = claim_element(nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div3 = claim_element(div6_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			h2 = claim_element(div3_nodes, "H2", { class: true });
    			var h2_nodes = children(h2);
    			t0 = claim_text(h2_nodes, "Delete all stored data");
    			h2_nodes.forEach(detach_dev);
    			t1 = claim_space(div3_nodes);

    			input = claim_element(div3_nodes, "INPUT", {
    				type: true,
    				placeholder: true,
    				class: true
    			});

    			t2 = claim_space(div3_nodes);
    			button = claim_element(div3_nodes, "BUTTON", { id: true, class: true });
    			var button_nodes = children(button);
    			t3 = claim_text(button_nodes, "Delete");
    			button_nodes.forEach(detach_dev);
    			t4 = claim_space(div3_nodes);
    			div0 = claim_element(div3_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t5 = claim_text(div0_nodes, /*message*/ ctx[2]);
    			div0_nodes.forEach(detach_dev);
    			t6 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			div1 = claim_element(div2_nodes, "DIV", { class: true, style: true });
    			children(div1).forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			t7 = claim_space(div6_nodes);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			t8 = claim_text(div4_nodes, "API Analytics");
    			div4_nodes.forEach(detach_dev);
    			t9 = claim_space(div5_nodes);
    			img = claim_element(div5_nodes, "IMG", { class: true, src: true, alt: true });
    			div5_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(h2, "class", "svelte-brqh97");
    			add_location(h2, file$4, 19, 4, 489);
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "Enter API key");
    			attr_dev(input, "class", "svelte-brqh97");
    			add_location(input, file$4, 20, 4, 526);
    			attr_dev(button, "id", "generateBtn");
    			attr_dev(button, "class", "svelte-brqh97");
    			add_location(button, file$4, 21, 4, 601);
    			attr_dev(div0, "class", "notification svelte-brqh97");
    			add_location(div0, file$4, 22, 4, 668);
    			attr_dev(div1, "class", "loader");
    			set_style(div1, "display", /*loading*/ ctx[1] ? 'initial' : 'none');
    			add_location(div1, file$4, 24, 6, 744);
    			attr_dev(div2, "class", "spinner");
    			add_location(div2, file$4, 23, 4, 715);
    			attr_dev(div3, "class", "content svelte-brqh97");
    			add_location(div3, file$4, 18, 2, 462);
    			attr_dev(div4, "class", "highlight logo svelte-brqh97");
    			add_location(div4, file$4, 29, 4, 947);
    			attr_dev(img, "class", "footer-logo");
    			if (!src_url_equal(img.src, img_src_value = "img/logo.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			add_location(img, file$4, 30, 4, 1000);
    			attr_dev(div5, "class", "details svelte-brqh97");
    			add_location(div5, file$4, 27, 2, 840);
    			attr_dev(div6, "class", "generate svelte-brqh97");
    			add_location(div6, file$4, 17, 0, 436);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div6, anchor);
    			append_hydration_dev(div6, div3);
    			append_hydration_dev(div3, h2);
    			append_hydration_dev(h2, t0);
    			append_hydration_dev(div3, t1);
    			append_hydration_dev(div3, input);
    			set_input_value(input, /*apiKey*/ ctx[0]);
    			append_hydration_dev(div3, t2);
    			append_hydration_dev(div3, button);
    			append_hydration_dev(button, t3);
    			append_hydration_dev(div3, t4);
    			append_hydration_dev(div3, div0);
    			append_hydration_dev(div0, t5);
    			append_hydration_dev(div3, t6);
    			append_hydration_dev(div3, div2);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div6, t7);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, t8);
    			append_hydration_dev(div5, t9);
    			append_hydration_dev(div5, img);

    			if (!mounted) {
    				dispose = [
    					listen_dev(input, "input", /*input_input_handler*/ ctx[4]),
    					listen_dev(button, "click", /*genAPIKey*/ ctx[3], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (dirty & /*apiKey*/ 1 && input.value !== /*apiKey*/ ctx[0]) {
    				set_input_value(input, /*apiKey*/ ctx[0]);
    			}

    			if (dirty & /*message*/ 4) set_data_dev(t5, /*message*/ ctx[2]);

    			if (dirty & /*loading*/ 2) {
    				set_style(div1, "display", /*loading*/ ctx[1] ? 'initial' : 'none');
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div6);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$5.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$5($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Delete', slots, []);
    	let apiKey = "";
    	let loading = false;
    	let message = "";

    	async function genAPIKey() {
    		$$invalidate(1, loading = true);

    		// Fetch page ID
    		const response = await fetch(`https://api-analytics-server.vercel.app/api/delete/${apiKey}`);

    		if (response.status == 200) {
    			$$invalidate(2, message = "Deleted successfully");
    		} else {
    			$$invalidate(2, message = "Error: API key invalid");
    		}

    		$$invalidate(1, loading = false);
    	}

    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Delete> was created with unknown prop '${key}'`);
    	});

    	function input_input_handler() {
    		apiKey = this.value;
    		$$invalidate(0, apiKey);
    	}

    	$$self.$capture_state = () => ({ apiKey, loading, message, genAPIKey });

    	$$self.$inject_state = $$props => {
    		if ('apiKey' in $$props) $$invalidate(0, apiKey = $$props.apiKey);
    		if ('loading' in $$props) $$invalidate(1, loading = $$props.loading);
    		if ('message' in $$props) $$invalidate(2, message = $$props.message);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [apiKey, loading, message, genAPIKey, input_input_handler];
    }

    class Delete extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$5, create_fragment$5, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Delete",
    			options,
    			id: create_fragment$5.name
    		});
    	}
    }

    /* src\components\monitoring\ResponseTime.svelte generated by Svelte v3.53.1 */
    const file$3 = "src\\components\\monitoring\\ResponseTime.svelte";

    function create_fragment$4(ctx) {
    	let div1;
    	let div0;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { id: true });
    			var div0_nodes = children(div0);
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "id", "plotDiv");
    			add_location(div0, file$3, 91, 2, 2190);
    			attr_dev(div1, "id", "plotly");
    			add_location(div1, file$3, 90, 0, 2169);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, div0);
    			/*div0_binding*/ ctx[4](div0);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    			/*div0_binding*/ ctx[4](null);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$4.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$4($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('ResponseTime', slots, []);

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
    		$$invalidate(3, setup = true);
    	});

    	let { data, period } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<ResponseTime> was created without expected prop 'data'");
    		}

    		if (period === undefined && !('period' in $$props || $$self.$$.bound[$$self.$$.props['period']])) {
    			console.warn("<ResponseTime> was created without expected prop 'period'");
    		}
    	});

    	const writable_props = ['data', 'period'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<ResponseTime> was created with unknown prop '${key}'`);
    	});

    	function div0_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			plotDiv = $$value;
    			$$invalidate(0, plotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		periodToMarkers: periodToMarkers$1,
    		defaultLayout,
    		bars,
    		buildPlotData,
    		genPlot,
    		plotDiv,
    		setup,
    		data,
    		period
    	});

    	$$self.$inject_state = $$props => {
    		if ('plotDiv' in $$props) $$invalidate(0, plotDiv = $$props.plotDiv);
    		if ('setup' in $$props) $$invalidate(3, setup = $$props.setup);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    		if ('period' in $$props) $$invalidate(2, period = $$props.period);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*period, setup*/ 12) {
    			period && setup && genPlot();
    		}
    	};

    	return [plotDiv, data, period, setup, div0_binding];
    }

    class ResponseTime extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$4, create_fragment$4, safe_not_equal, { data: 1, period: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "ResponseTime",
    			options,
    			id: create_fragment$4.name
    		});
    	}

    	get data() {
    		throw new Error("<ResponseTime>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<ResponseTime>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get period() {
    		throw new Error("<ResponseTime>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set period(value) {
    		throw new Error("<ResponseTime>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\monitoring\Card.svelte generated by Svelte v3.53.1 */
    const file$2 = "src\\components\\monitoring\\Card.svelte";

    function get_each_context(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[10] = list[i];
    	return child_ctx;
    }

    // (71:8) {:else}
    function create_else_block$1(ctx) {
    	let img;
    	let img_src_value;

    	const block = {
    		c: function create() {
    			img = element("img");
    			this.h();
    		},
    		l: function claim(nodes) {
    			img = claim_element(nodes, "IMG", { src: true, alt: true });
    			this.h();
    		},
    		h: function hydrate() {
    			if (!src_url_equal(img.src, img_src_value = "/img/smalltick.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			add_location(img, file$2, 71, 10, 1812);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, img, anchor);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(img);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_else_block$1.name,
    		type: "else",
    		source: "(71:8) {:else}",
    		ctx
    	});

    	return block;
    }

    // (69:8) {#if error}
    function create_if_block$1(ctx) {
    	let img;
    	let img_src_value;

    	const block = {
    		c: function create() {
    			img = element("img");
    			this.h();
    		},
    		l: function claim(nodes) {
    			img = claim_element(nodes, "IMG", { src: true, alt: true });
    			this.h();
    		},
    		h: function hydrate() {
    			if (!src_url_equal(img.src, img_src_value = "/img/smallcross.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			add_location(img, file$2, 69, 10, 1743);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, img, anchor);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(img);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block$1.name,
    		type: "if",
    		source: "(69:8) {#if error}",
    		ctx
    	});

    	return block;
    }

    // (82:4) {#each measurements as measurement}
    function create_each_block(ctx) {
    	let div;
    	let div_class_value;

    	const block = {
    		c: function create() {
    			div = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true });
    			children(div).forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", div_class_value = "measurement " + /*measurement*/ ctx[10].status + " svelte-14ep2gi");
    			add_location(div, file$2, 82, 6, 2126);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*measurements*/ 16 && div_class_value !== (div_class_value = "measurement " + /*measurement*/ ctx[10].status + " svelte-14ep2gi")) {
    				attr_dev(div, "class", div_class_value);
    			}
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block.name,
    		type: "each",
    		source: "(82:4) {#each measurements as measurement}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$3(ctx) {
    	let div8;
    	let div5;
    	let div2;
    	let div0;
    	let t0;
    	let div1;
    	let t1_value = /*data*/ ctx[0].name + "";
    	let t1;
    	let t2;
    	let div4;
    	let div3;
    	let t3;
    	let t4;
    	let t5;
    	let t6;
    	let div6;
    	let t7;
    	let div7;
    	let responsetime;
    	let current;

    	function select_block_type(ctx, dirty) {
    		if (/*error*/ ctx[3]) return create_if_block$1;
    		return create_else_block$1;
    	}

    	let current_block_type = select_block_type(ctx);
    	let if_block = current_block_type(ctx);
    	let each_value = /*measurements*/ ctx[4];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block(get_each_context(ctx, each_value, i));
    	}

    	responsetime = new ResponseTime({
    			props: {
    				data: /*measurements*/ ctx[4],
    				period: /*period*/ ctx[1]
    			},
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			div8 = element("div");
    			div5 = element("div");
    			div2 = element("div");
    			div0 = element("div");
    			if_block.c();
    			t0 = space();
    			div1 = element("div");
    			t1 = text(t1_value);
    			t2 = space();
    			div4 = element("div");
    			div3 = element("div");
    			t3 = text("Uptime: ");
    			t4 = text(/*uptime*/ ctx[2]);
    			t5 = text("%");
    			t6 = space();
    			div6 = element("div");

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].c();
    			}

    			t7 = space();
    			div7 = element("div");
    			create_component(responsetime.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div8 = claim_element(nodes, "DIV", { class: true });
    			var div8_nodes = children(div8);
    			div5 = claim_element(div8_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			div2 = claim_element(div5_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			div0 = claim_element(div2_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			if_block.l(div0_nodes);
    			div0_nodes.forEach(detach_dev);
    			t0 = claim_space(div2_nodes);
    			div1 = claim_element(div2_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			t1 = claim_text(div1_nodes, t1_value);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			t2 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			div3 = claim_element(div4_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			t3 = claim_text(div3_nodes, "Uptime: ");
    			t4 = claim_text(div3_nodes, /*uptime*/ ctx[2]);
    			t5 = claim_text(div3_nodes, "%");
    			div3_nodes.forEach(detach_dev);
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			t6 = claim_space(div8_nodes);
    			div6 = claim_element(div8_nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].l(div6_nodes);
    			}

    			div6_nodes.forEach(detach_dev);
    			t7 = claim_space(div8_nodes);
    			div7 = claim_element(div8_nodes, "DIV", { class: true });
    			var div7_nodes = children(div7);
    			claim_component(responsetime.$$.fragment, div7_nodes);
    			div7_nodes.forEach(detach_dev);
    			div8_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-status");
    			add_location(div0, file$2, 67, 6, 1685);
    			attr_dev(div1, "class", "endpoint svelte-14ep2gi");
    			add_location(div1, file$2, 74, 6, 1888);
    			attr_dev(div2, "class", "card-text-left svelte-14ep2gi");
    			add_location(div2, file$2, 66, 4, 1649);
    			attr_dev(div3, "class", "uptime svelte-14ep2gi");
    			add_location(div3, file$2, 77, 6, 1982);
    			attr_dev(div4, "class", "card-text-right");
    			add_location(div4, file$2, 76, 4, 1945);
    			attr_dev(div5, "class", "card-text svelte-14ep2gi");
    			add_location(div5, file$2, 65, 2, 1620);
    			attr_dev(div6, "class", "measurements svelte-14ep2gi");
    			add_location(div6, file$2, 80, 2, 2051);
    			attr_dev(div7, "class", "response-time");
    			add_location(div7, file$2, 85, 2, 2201);
    			attr_dev(div8, "class", "card svelte-14ep2gi");
    			toggle_class(div8, "card-error", /*error*/ ctx[3]);
    			add_location(div8, file$2, 64, 0, 1573);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div8, anchor);
    			append_hydration_dev(div8, div5);
    			append_hydration_dev(div5, div2);
    			append_hydration_dev(div2, div0);
    			if_block.m(div0, null);
    			append_hydration_dev(div2, t0);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div1, t1);
    			append_hydration_dev(div5, t2);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, div3);
    			append_hydration_dev(div3, t3);
    			append_hydration_dev(div3, t4);
    			append_hydration_dev(div3, t5);
    			append_hydration_dev(div8, t6);
    			append_hydration_dev(div8, div6);

    			for (let i = 0; i < each_blocks.length; i += 1) {
    				each_blocks[i].m(div6, null);
    			}

    			append_hydration_dev(div8, t7);
    			append_hydration_dev(div8, div7);
    			mount_component(responsetime, div7, null);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			if (current_block_type !== (current_block_type = select_block_type(ctx))) {
    				if_block.d(1);
    				if_block = current_block_type(ctx);

    				if (if_block) {
    					if_block.c();
    					if_block.m(div0, null);
    				}
    			}

    			if ((!current || dirty & /*data*/ 1) && t1_value !== (t1_value = /*data*/ ctx[0].name + "")) set_data_dev(t1, t1_value);
    			if (!current || dirty & /*uptime*/ 4) set_data_dev(t4, /*uptime*/ ctx[2]);

    			if (dirty & /*measurements*/ 16) {
    				each_value = /*measurements*/ ctx[4];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block(child_ctx);
    						each_blocks[i].c();
    						each_blocks[i].m(div6, null);
    					}
    				}

    				for (; i < each_blocks.length; i += 1) {
    					each_blocks[i].d(1);
    				}

    				each_blocks.length = each_value.length;
    			}

    			const responsetime_changes = {};
    			if (dirty & /*measurements*/ 16) responsetime_changes.data = /*measurements*/ ctx[4];
    			if (dirty & /*period*/ 2) responsetime_changes.period = /*period*/ ctx[1];
    			responsetime.$set(responsetime_changes);

    			if (!current || dirty & /*error*/ 8) {
    				toggle_class(div8, "card-error", /*error*/ ctx[3]);
    			}
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(responsetime.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(responsetime.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div8);
    			if_block.d();
    			destroy_each(each_blocks, detaching);
    			destroy_component(responsetime);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$3.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

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

    function instance$3($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Card', slots, []);

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
    			$$invalidate(2, uptime = "100");
    		} else {
    			$$invalidate(2, uptime = per.toFixed(1));
    		}
    	}

    	function setMeasurements() {
    		let markers = periodToMarkers(period);
    		$$invalidate(4, measurements = Array(markers).fill({ status: null, response_time: 0 }));
    		let start = markers - data.measurements.length;

    		for (let i = 0; i < data.measurements.length; i++) {
    			$$invalidate(4, measurements[i + start] = data.measurements[i], measurements);
    		}
    	}

    	function setError() {
    		$$invalidate(3, error = measurements[measurements.length - 1].status == "error");
    		$$invalidate(5, anyError = anyError || error);
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

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Card> was created without expected prop 'data'");
    		}

    		if (period === undefined && !('period' in $$props || $$self.$$.bound[$$self.$$.props['period']])) {
    			console.warn("<Card> was created without expected prop 'period'");
    		}

    		if (anyError === undefined && !('anyError' in $$props || $$self.$$.bound[$$self.$$.props['anyError']])) {
    			console.warn("<Card> was created without expected prop 'anyError'");
    		}
    	});

    	const writable_props = ['data', 'period', 'anyError'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Card> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    		if ('period' in $$props) $$invalidate(1, period = $$props.period);
    		if ('anyError' in $$props) $$invalidate(5, anyError = $$props.anyError);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		ResponseTime,
    		setUptime,
    		periodToMarkers,
    		setMeasurements,
    		setError,
    		build,
    		uptime,
    		error,
    		measurements,
    		data,
    		period,
    		anyError
    	});

    	$$self.$inject_state = $$props => {
    		if ('uptime' in $$props) $$invalidate(2, uptime = $$props.uptime);
    		if ('error' in $$props) $$invalidate(3, error = $$props.error);
    		if ('measurements' in $$props) $$invalidate(4, measurements = $$props.measurements);
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    		if ('period' in $$props) $$invalidate(1, period = $$props.period);
    		if ('anyError' in $$props) $$invalidate(5, anyError = $$props.anyError);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	$$self.$$.update = () => {
    		if ($$self.$$.dirty & /*period*/ 2) {
    			period && build();
    		}
    	};

    	return [data, period, uptime, error, measurements, anyError];
    }

    class Card extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$3, create_fragment$3, safe_not_equal, { data: 0, period: 1, anyError: 5 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Card",
    			options,
    			id: create_fragment$3.name
    		});
    	}

    	get data() {
    		throw new Error("<Card>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Card>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get period() {
    		throw new Error("<Card>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set period(value) {
    		throw new Error("<Card>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	get anyError() {
    		throw new Error("<Card>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set anyError(value) {
    		throw new Error("<Card>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\monitoring\TrackNew.svelte generated by Svelte v3.53.1 */

    const file$1 = "src\\components\\monitoring\\TrackNew.svelte";

    function create_fragment$2(ctx) {
    	let div4;
    	let div3;
    	let div1;
    	let div0;
    	let t0;
    	let t1;
    	let input;
    	let t2;
    	let button;
    	let t3;
    	let t4;
    	let div2;
    	let t5;
    	let b0;
    	let t6;
    	let t7;
    	let b1;
    	let t8;
    	let t9;

    	const block = {
    		c: function create() {
    			div4 = element("div");
    			div3 = element("div");
    			div1 = element("div");
    			div0 = element("div");
    			t0 = text("URL");
    			t1 = space();
    			input = element("input");
    			t2 = space();
    			button = element("button");
    			t3 = text("Add");
    			t4 = space();
    			div2 = element("div");
    			t5 = text("Endpoints are pinged by our servers every 30 mins and response ");
    			b0 = element("b");
    			t6 = text("status");
    			t7 = text("\r\n      and response ");
    			b1 = element("b");
    			t8 = text("time");
    			t9 = text(" are logged.");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div4 = claim_element(nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			div3 = claim_element(div4_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			div1 = claim_element(div3_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "URL");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div1_nodes);
    			input = claim_element(div1_nodes, "INPUT", { type: true, class: true });
    			t2 = claim_space(div1_nodes);
    			button = claim_element(div1_nodes, "BUTTON", { class: true });
    			var button_nodes = children(button);
    			t3 = claim_text(button_nodes, "Add");
    			button_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			t4 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			t5 = claim_text(div2_nodes, "Endpoints are pinged by our servers every 30 mins and response ");
    			b0 = claim_element(div2_nodes, "B", {});
    			var b0_nodes = children(b0);
    			t6 = claim_text(b0_nodes, "status");
    			b0_nodes.forEach(detach_dev);
    			t7 = claim_text(div2_nodes, "\r\n      and response ");
    			b1 = claim_element(div2_nodes, "B", {});
    			var b1_nodes = children(b1);
    			t8 = claim_text(b1_nodes, "time");
    			b1_nodes.forEach(detach_dev);
    			t9 = claim_text(div2_nodes, " are logged.");
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			div4_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "url-text svelte-m15jzh");
    			add_location(div0, file$1, 5, 6, 107);
    			attr_dev(input, "type", "text");
    			attr_dev(input, "class", "svelte-m15jzh");
    			add_location(input, file$1, 6, 6, 146);
    			attr_dev(button, "class", "svelte-m15jzh");
    			add_location(button, file$1, 7, 6, 175);
    			attr_dev(div1, "class", "url svelte-m15jzh");
    			add_location(div1, file$1, 4, 4, 82);
    			add_location(b0, file$1, 10, 69, 304);
    			add_location(b1, file$1, 13, 19, 356);
    			attr_dev(div2, "class", "detail svelte-m15jzh");
    			add_location(div2, file$1, 9, 4, 213);
    			attr_dev(div3, "class", "card-text svelte-m15jzh");
    			add_location(div3, file$1, 3, 2, 53);
    			attr_dev(div4, "class", "card svelte-m15jzh");
    			add_location(div4, file$1, 2, 0, 31);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div4, anchor);
    			append_hydration_dev(div4, div3);
    			append_hydration_dev(div3, div1);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div1, t1);
    			append_hydration_dev(div1, input);
    			append_hydration_dev(div1, t2);
    			append_hydration_dev(div1, button);
    			append_hydration_dev(button, t3);
    			append_hydration_dev(div3, t4);
    			append_hydration_dev(div3, div2);
    			append_hydration_dev(div2, t5);
    			append_hydration_dev(div2, b0);
    			append_hydration_dev(b0, t6);
    			append_hydration_dev(div2, t7);
    			append_hydration_dev(div2, b1);
    			append_hydration_dev(b1, t8);
    			append_hydration_dev(div2, t9);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div4);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$2.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$2($$self, $$props) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('TrackNew', slots, []);
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<TrackNew> was created with unknown prop '${key}'`);
    	});

    	return [];
    }

    class TrackNew extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$2, create_fragment$2, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "TrackNew",
    			options,
    			id: create_fragment$2.name
    		});
    	}
    }

    /* src\routes\Monitoring.svelte generated by Svelte v3.53.1 */

    const { console: console_1 } = globals;
    const file = "src\\routes\\Monitoring.svelte";

    // (65:4) {:else}
    function create_else_block(ctx) {
    	let div1;
    	let img;
    	let img_src_value;
    	let t0;
    	let div0;
    	let t1;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			img = element("img");
    			t0 = space();
    			div0 = element("div");
    			t1 = text("Systems Online");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);

    			img = claim_element(div1_nodes, "IMG", {
    				id: true,
    				src: true,
    				alt: true,
    				class: true
    			});

    			t0 = claim_space(div1_nodes);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t1 = claim_text(div0_nodes, "Systems Online");
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(img, "id", "status-image");
    			if (!src_url_equal(img.src, img_src_value = "/img/bigtick.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			attr_dev(img, "class", "svelte-11gf7xd");
    			add_location(img, file, 66, 8, 1984);
    			attr_dev(div0, "class", "status-text svelte-11gf7xd");
    			add_location(div0, file, 67, 8, 2049);
    			attr_dev(div1, "class", "status-image");
    			add_location(div1, file, 65, 6, 1948);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, img);
    			append_hydration_dev(div1, t0);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t1);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_else_block.name,
    		type: "else",
    		source: "(65:4) {:else}",
    		ctx
    	});

    	return block;
    }

    // (60:4) {#if error}
    function create_if_block_1(ctx) {
    	let div1;
    	let img;
    	let img_src_value;
    	let t0;
    	let div0;
    	let t1;

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			img = element("img");
    			t0 = space();
    			div0 = element("div");
    			t1 = text("Systems down");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);

    			img = claim_element(div1_nodes, "IMG", {
    				id: true,
    				src: true,
    				alt: true,
    				class: true
    			});

    			t0 = claim_space(div1_nodes);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t1 = claim_text(div0_nodes, "Systems down");
    			div0_nodes.forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(img, "id", "status-image");
    			if (!src_url_equal(img.src, img_src_value = "/img/bigcross.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			attr_dev(img, "class", "svelte-11gf7xd");
    			add_location(img, file, 61, 8, 1804);
    			attr_dev(div0, "class", "status-text svelte-11gf7xd");
    			add_location(div0, file, 62, 8, 1870);
    			attr_dev(div1, "class", "status-image");
    			add_location(div1, file, 60, 6, 1768);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div1, anchor);
    			append_hydration_dev(div1, img);
    			append_hydration_dev(div1, t0);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div0, t1);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div1);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(60:4) {#if error}",
    		ctx
    	});

    	return block;
    }

    // (118:4) {#if showTrackNew || measurements.length == 0}
    function create_if_block(ctx) {
    	let tracknew;
    	let current;
    	tracknew = new TrackNew({ $$inline: true });

    	const block = {
    		c: function create() {
    			create_component(tracknew.$$.fragment);
    		},
    		l: function claim(nodes) {
    			claim_component(tracknew.$$.fragment, nodes);
    		},
    		m: function mount(target, anchor) {
    			mount_component(tracknew, target, anchor);
    			current = true;
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(tracknew.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(tracknew.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(tracknew, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block.name,
    		type: "if",
    		source: "(118:4) {#if showTrackNew || measurements.length == 0}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$1(ctx) {
    	let div7;
    	let div0;
    	let t0;
    	let div6;
    	let div5;
    	let div2;
    	let button0;
    	let div1;
    	let span;
    	let t1;
    	let t2;
    	let t3;
    	let div4;
    	let div3;
    	let button1;
    	let t4;
    	let button1_class_value;
    	let t5;
    	let button2;
    	let t6;
    	let button2_class_value;
    	let t7;
    	let button3;
    	let t8;
    	let button3_class_value;
    	let t9;
    	let button4;
    	let t10;
    	let button4_class_value;
    	let t11;
    	let t12;
    	let card0;
    	let updating_anyError;
    	let t13;
    	let card1;
    	let updating_anyError_1;
    	let t14;
    	let card2;
    	let updating_anyError_2;
    	let t15;
    	let footer;
    	let current;
    	let mounted;
    	let dispose;

    	function select_block_type(ctx, dirty) {
    		if (/*error*/ ctx[0]) return create_if_block_1;
    		return create_else_block;
    	}

    	let current_block_type = select_block_type(ctx);
    	let if_block0 = current_block_type(ctx);
    	let if_block1 = (/*showTrackNew*/ ctx[3] || /*measurements*/ ctx[2].length == 0) && create_if_block(ctx);

    	function card0_anyError_binding(value) {
    		/*card0_anyError_binding*/ ctx[11](value);
    	}

    	let card0_props = {
    		data: /*measurements*/ ctx[2][0],
    		period: /*period*/ ctx[1]
    	};

    	if (/*error*/ ctx[0] !== void 0) {
    		card0_props.anyError = /*error*/ ctx[0];
    	}

    	card0 = new Card({ props: card0_props, $$inline: true });
    	binding_callbacks.push(() => bind(card0, 'anyError', card0_anyError_binding));

    	function card1_anyError_binding(value) {
    		/*card1_anyError_binding*/ ctx[12](value);
    	}

    	let card1_props = {
    		data: /*measurements*/ ctx[2][1],
    		period: /*period*/ ctx[1]
    	};

    	if (/*error*/ ctx[0] !== void 0) {
    		card1_props.anyError = /*error*/ ctx[0];
    	}

    	card1 = new Card({ props: card1_props, $$inline: true });
    	binding_callbacks.push(() => bind(card1, 'anyError', card1_anyError_binding));

    	function card2_anyError_binding(value) {
    		/*card2_anyError_binding*/ ctx[13](value);
    	}

    	let card2_props = {
    		data: /*measurements*/ ctx[2][2],
    		period: /*period*/ ctx[1]
    	};

    	if (/*error*/ ctx[0] !== void 0) {
    		card2_props.anyError = /*error*/ ctx[0];
    	}

    	card2 = new Card({ props: card2_props, $$inline: true });
    	binding_callbacks.push(() => bind(card2, 'anyError', card2_anyError_binding));
    	footer = new Footer({ $$inline: true });

    	const block = {
    		c: function create() {
    			div7 = element("div");
    			div0 = element("div");
    			if_block0.c();
    			t0 = space();
    			div6 = element("div");
    			div5 = element("div");
    			div2 = element("div");
    			button0 = element("button");
    			div1 = element("div");
    			span = element("span");
    			t1 = text("+");
    			t2 = text(" New");
    			t3 = space();
    			div4 = element("div");
    			div3 = element("div");
    			button1 = element("button");
    			t4 = text("24h");
    			t5 = space();
    			button2 = element("button");
    			t6 = text("7d");
    			t7 = space();
    			button3 = element("button");
    			t8 = text("30d");
    			t9 = space();
    			button4 = element("button");
    			t10 = text("60d");
    			t11 = space();
    			if (if_block1) if_block1.c();
    			t12 = space();
    			create_component(card0.$$.fragment);
    			t13 = space();
    			create_component(card1.$$.fragment);
    			t14 = space();
    			create_component(card2.$$.fragment);
    			t15 = space();
    			create_component(footer.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div7 = claim_element(nodes, "DIV", { class: true });
    			var div7_nodes = children(div7);
    			div0 = claim_element(div7_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			if_block0.l(div0_nodes);
    			div0_nodes.forEach(detach_dev);
    			t0 = claim_space(div7_nodes);
    			div6 = claim_element(div7_nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			div2 = claim_element(div5_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			button0 = claim_element(div2_nodes, "BUTTON", { class: true });
    			var button0_nodes = children(button0);
    			div1 = claim_element(button0_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			span = claim_element(div1_nodes, "SPAN", { class: true });
    			var span_nodes = children(span);
    			t1 = claim_text(span_nodes, "+");
    			span_nodes.forEach(detach_dev);
    			t2 = claim_text(div1_nodes, " New");
    			div1_nodes.forEach(detach_dev);
    			button0_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			t3 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			div3 = claim_element(div4_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			button1 = claim_element(div3_nodes, "BUTTON", { class: true });
    			var button1_nodes = children(button1);
    			t4 = claim_text(button1_nodes, "24h");
    			button1_nodes.forEach(detach_dev);
    			t5 = claim_space(div3_nodes);
    			button2 = claim_element(div3_nodes, "BUTTON", { class: true });
    			var button2_nodes = children(button2);
    			t6 = claim_text(button2_nodes, "7d");
    			button2_nodes.forEach(detach_dev);
    			t7 = claim_space(div3_nodes);
    			button3 = claim_element(div3_nodes, "BUTTON", { class: true });
    			var button3_nodes = children(button3);
    			t8 = claim_text(button3_nodes, "30d");
    			button3_nodes.forEach(detach_dev);
    			t9 = claim_space(div3_nodes);
    			button4 = claim_element(div3_nodes, "BUTTON", { class: true });
    			var button4_nodes = children(button4);
    			t10 = claim_text(button4_nodes, "60d");
    			button4_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			t11 = claim_space(div6_nodes);
    			if (if_block1) if_block1.l(div6_nodes);
    			t12 = claim_space(div6_nodes);
    			claim_component(card0.$$.fragment, div6_nodes);
    			t13 = claim_space(div6_nodes);
    			claim_component(card1.$$.fragment, div6_nodes);
    			t14 = claim_space(div6_nodes);
    			claim_component(card2.$$.fragment, div6_nodes);
    			div6_nodes.forEach(detach_dev);
    			div7_nodes.forEach(detach_dev);
    			t15 = claim_space(nodes);
    			claim_component(footer.$$.fragment, nodes);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "status svelte-11gf7xd");
    			add_location(div0, file, 58, 2, 1723);
    			attr_dev(span, "class", "plus svelte-11gf7xd");
    			add_location(span, file, 76, 12, 2339);
    			attr_dev(div1, "class", "add-new-text svelte-11gf7xd");
    			add_location(div1, file, 75, 11, 2299);
    			attr_dev(button0, "class", "add-new-btn svelte-11gf7xd");
    			add_location(button0, file, 74, 8, 2229);
    			attr_dev(div2, "class", "add-new svelte-11gf7xd");
    			add_location(div2, file, 73, 6, 2198);
    			attr_dev(button1, "class", button1_class_value = "period-btn " + (/*period*/ ctx[1] == '24h' ? 'active' : '') + " svelte-11gf7xd");
    			add_location(button1, file, 82, 10, 2519);
    			attr_dev(button2, "class", button2_class_value = "period-btn " + (/*period*/ ctx[1] == '7d' ? 'active' : '') + " svelte-11gf7xd");
    			add_location(button2, file, 90, 10, 2735);
    			attr_dev(button3, "class", button3_class_value = "period-btn " + (/*period*/ ctx[1] == '30d' ? 'active' : '') + " svelte-11gf7xd");
    			add_location(button3, file, 98, 10, 2948);
    			attr_dev(button4, "class", button4_class_value = "period-btn " + (/*period*/ ctx[1] == '60d' ? 'active' : '') + " svelte-11gf7xd");
    			add_location(button4, file, 106, 10, 3164);
    			attr_dev(div3, "class", "period-controls svelte-11gf7xd");
    			add_location(div3, file, 81, 8, 2478);
    			attr_dev(div4, "class", "period-controls-container");
    			add_location(div4, file, 80, 6, 2429);
    			attr_dev(div5, "class", "controls svelte-11gf7xd");
    			add_location(div5, file, 72, 4, 2168);
    			attr_dev(div6, "class", "cards-container svelte-11gf7xd");
    			add_location(div6, file, 71, 2, 2133);
    			attr_dev(div7, "class", "monitoring svelte-11gf7xd");
    			add_location(div7, file, 57, 0, 1695);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div7, anchor);
    			append_hydration_dev(div7, div0);
    			if_block0.m(div0, null);
    			append_hydration_dev(div7, t0);
    			append_hydration_dev(div7, div6);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, div2);
    			append_hydration_dev(div2, button0);
    			append_hydration_dev(button0, div1);
    			append_hydration_dev(div1, span);
    			append_hydration_dev(span, t1);
    			append_hydration_dev(div1, t2);
    			append_hydration_dev(div5, t3);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, div3);
    			append_hydration_dev(div3, button1);
    			append_hydration_dev(button1, t4);
    			append_hydration_dev(div3, t5);
    			append_hydration_dev(div3, button2);
    			append_hydration_dev(button2, t6);
    			append_hydration_dev(div3, t7);
    			append_hydration_dev(div3, button3);
    			append_hydration_dev(button3, t8);
    			append_hydration_dev(div3, t9);
    			append_hydration_dev(div3, button4);
    			append_hydration_dev(button4, t10);
    			append_hydration_dev(div6, t11);
    			if (if_block1) if_block1.m(div6, null);
    			append_hydration_dev(div6, t12);
    			mount_component(card0, div6, null);
    			append_hydration_dev(div6, t13);
    			mount_component(card1, div6, null);
    			append_hydration_dev(div6, t14);
    			mount_component(card2, div6, null);
    			insert_hydration_dev(target, t15, anchor);
    			mount_component(footer, target, anchor);
    			current = true;

    			if (!mounted) {
    				dispose = [
    					listen_dev(button0, "click", /*toggleShowTrackNew*/ ctx[5], false, false, false),
    					listen_dev(button1, "click", /*click_handler*/ ctx[7], false, false, false),
    					listen_dev(button2, "click", /*click_handler_1*/ ctx[8], false, false, false),
    					listen_dev(button3, "click", /*click_handler_2*/ ctx[9], false, false, false),
    					listen_dev(button4, "click", /*click_handler_3*/ ctx[10], false, false, false)
    				];

    				mounted = true;
    			}
    		},
    		p: function update(ctx, [dirty]) {
    			if (current_block_type !== (current_block_type = select_block_type(ctx))) {
    				if_block0.d(1);
    				if_block0 = current_block_type(ctx);

    				if (if_block0) {
    					if_block0.c();
    					if_block0.m(div0, null);
    				}
    			}

    			if (!current || dirty & /*period*/ 2 && button1_class_value !== (button1_class_value = "period-btn " + (/*period*/ ctx[1] == '24h' ? 'active' : '') + " svelte-11gf7xd")) {
    				attr_dev(button1, "class", button1_class_value);
    			}

    			if (!current || dirty & /*period*/ 2 && button2_class_value !== (button2_class_value = "period-btn " + (/*period*/ ctx[1] == '7d' ? 'active' : '') + " svelte-11gf7xd")) {
    				attr_dev(button2, "class", button2_class_value);
    			}

    			if (!current || dirty & /*period*/ 2 && button3_class_value !== (button3_class_value = "period-btn " + (/*period*/ ctx[1] == '30d' ? 'active' : '') + " svelte-11gf7xd")) {
    				attr_dev(button3, "class", button3_class_value);
    			}

    			if (!current || dirty & /*period*/ 2 && button4_class_value !== (button4_class_value = "period-btn " + (/*period*/ ctx[1] == '60d' ? 'active' : '') + " svelte-11gf7xd")) {
    				attr_dev(button4, "class", button4_class_value);
    			}

    			if (/*showTrackNew*/ ctx[3] || /*measurements*/ ctx[2].length == 0) {
    				if (if_block1) {
    					if (dirty & /*showTrackNew, measurements*/ 12) {
    						transition_in(if_block1, 1);
    					}
    				} else {
    					if_block1 = create_if_block(ctx);
    					if_block1.c();
    					transition_in(if_block1, 1);
    					if_block1.m(div6, t12);
    				}
    			} else if (if_block1) {
    				group_outros();

    				transition_out(if_block1, 1, 1, () => {
    					if_block1 = null;
    				});

    				check_outros();
    			}

    			const card0_changes = {};
    			if (dirty & /*measurements*/ 4) card0_changes.data = /*measurements*/ ctx[2][0];
    			if (dirty & /*period*/ 2) card0_changes.period = /*period*/ ctx[1];

    			if (!updating_anyError && dirty & /*error*/ 1) {
    				updating_anyError = true;
    				card0_changes.anyError = /*error*/ ctx[0];
    				add_flush_callback(() => updating_anyError = false);
    			}

    			card0.$set(card0_changes);
    			const card1_changes = {};
    			if (dirty & /*measurements*/ 4) card1_changes.data = /*measurements*/ ctx[2][1];
    			if (dirty & /*period*/ 2) card1_changes.period = /*period*/ ctx[1];

    			if (!updating_anyError_1 && dirty & /*error*/ 1) {
    				updating_anyError_1 = true;
    				card1_changes.anyError = /*error*/ ctx[0];
    				add_flush_callback(() => updating_anyError_1 = false);
    			}

    			card1.$set(card1_changes);
    			const card2_changes = {};
    			if (dirty & /*measurements*/ 4) card2_changes.data = /*measurements*/ ctx[2][2];
    			if (dirty & /*period*/ 2) card2_changes.period = /*period*/ ctx[1];

    			if (!updating_anyError_2 && dirty & /*error*/ 1) {
    				updating_anyError_2 = true;
    				card2_changes.anyError = /*error*/ ctx[0];
    				add_flush_callback(() => updating_anyError_2 = false);
    			}

    			card2.$set(card2_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(if_block1);
    			transition_in(card0.$$.fragment, local);
    			transition_in(card1.$$.fragment, local);
    			transition_in(card2.$$.fragment, local);
    			transition_in(footer.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block1);
    			transition_out(card0.$$.fragment, local);
    			transition_out(card1.$$.fragment, local);
    			transition_out(card2.$$.fragment, local);
    			transition_out(footer.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div7);
    			if_block0.d();
    			if (if_block1) if_block1.d();
    			destroy_component(card0);
    			destroy_component(card1);
    			destroy_component(card2);
    			if (detaching) detach_dev(t15);
    			destroy_component(footer, detaching);
    			mounted = false;
    			run_all(dispose);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment$1.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function formatUUID(userID) {
    	return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(12, 16)}-${userID.slice(16, 20)}-${userID.slice(20)}`;
    }

    function instance$1($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Monitoring', slots, []);

    	async function fetchData() {
    		$$invalidate(6, userID = formatUUID(userID));

    		// Fetch page ID
    		try {
    			const response = await fetch(`https://api-analytics-server.vercel.app/api/user-data/${userID}`);

    			if (response.status == 200) {
    				const json = await response.json();
    				data = json.value;
    				console.log(data);
    			}
    		} catch(e) {
    			failed = true;
    		}
    	}

    	function setPeriod(value) {
    		$$invalidate(1, period = value);
    		$$invalidate(0, error = false);
    	}

    	function toggleShowTrackNew() {
    		$$invalidate(3, showTrackNew = !showTrackNew);
    	}

    	let error = false;
    	let period = "30d";
    	let data;
    	let measurements = Array(3);
    	let failed = false;

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
    	let showTrackNew = false;

    	onMount(() => {
    		fetchData();
    	});

    	let { userID } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (userID === undefined && !('userID' in $$props || $$self.$$.bound[$$self.$$.props['userID']])) {
    			console_1.warn("<Monitoring> was created without expected prop 'userID'");
    		}
    	});

    	const writable_props = ['userID'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console_1.warn(`<Monitoring> was created with unknown prop '${key}'`);
    	});

    	const click_handler = () => {
    		setPeriod("24h");
    	};

    	const click_handler_1 = () => {
    		setPeriod("7d");
    	};

    	const click_handler_2 = () => {
    		setPeriod("30d");
    	};

    	const click_handler_3 = () => {
    		setPeriod("60d");
    	};

    	function card0_anyError_binding(value) {
    		error = value;
    		$$invalidate(0, error);
    	}

    	function card1_anyError_binding(value) {
    		error = value;
    		$$invalidate(0, error);
    	}

    	function card2_anyError_binding(value) {
    		error = value;
    		$$invalidate(0, error);
    	}

    	$$self.$$set = $$props => {
    		if ('userID' in $$props) $$invalidate(6, userID = $$props.userID);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		Footer,
    		Card,
    		TrackNew,
    		formatUUID,
    		fetchData,
    		setPeriod,
    		toggleShowTrackNew,
    		error,
    		period,
    		data,
    		measurements,
    		failed,
    		showTrackNew,
    		userID
    	});

    	$$self.$inject_state = $$props => {
    		if ('error' in $$props) $$invalidate(0, error = $$props.error);
    		if ('period' in $$props) $$invalidate(1, period = $$props.period);
    		if ('data' in $$props) data = $$props.data;
    		if ('measurements' in $$props) $$invalidate(2, measurements = $$props.measurements);
    		if ('failed' in $$props) failed = $$props.failed;
    		if ('showTrackNew' in $$props) $$invalidate(3, showTrackNew = $$props.showTrackNew);
    		if ('userID' in $$props) $$invalidate(6, userID = $$props.userID);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		error,
    		period,
    		measurements,
    		showTrackNew,
    		setPeriod,
    		toggleShowTrackNew,
    		userID,
    		click_handler,
    		click_handler_1,
    		click_handler_2,
    		click_handler_3,
    		card0_anyError_binding,
    		card1_anyError_binding,
    		card2_anyError_binding
    	];
    }

    class Monitoring extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$1, create_fragment$1, safe_not_equal, { userID: 6 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Monitoring",
    			options,
    			id: create_fragment$1.name
    		});
    	}

    	get userID() {
    		throw new Error("<Monitoring>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set userID(value) {
    		throw new Error("<Monitoring>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\App.svelte generated by Svelte v3.53.1 */

    // (12:0) <Router {url}>
    function create_default_slot(ctx) {
    	let route0;
    	let t0;
    	let route1;
    	let t1;
    	let route2;
    	let t2;
    	let route3;
    	let t3;
    	let route4;
    	let t4;
    	let route5;
    	let t5;
    	let route6;
    	let current;

    	route0 = new Route({
    			props: { path: "/", component: Home },
    			$$inline: true
    		});

    	route1 = new Route({
    			props: { path: "/generate", component: Generate },
    			$$inline: true
    		});

    	route2 = new Route({
    			props: { path: "/dashboard", component: SignIn },
    			$$inline: true
    		});

    	route3 = new Route({
    			props: {
    				path: "/dashboard/demo",
    				component: Dashboard,
    				demo: true,
    				userID: null
    			},
    			$$inline: true
    		});

    	route4 = new Route({
    			props: {
    				path: "/dashboard/:userID",
    				component: Dashboard,
    				demo: false
    			},
    			$$inline: true
    		});

    	route5 = new Route({
    			props: {
    				path: "/monitoring/:userID",
    				component: Monitoring
    			},
    			$$inline: true
    		});

    	route6 = new Route({
    			props: { path: "/delete", component: Delete },
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			create_component(route0.$$.fragment);
    			t0 = space();
    			create_component(route1.$$.fragment);
    			t1 = space();
    			create_component(route2.$$.fragment);
    			t2 = space();
    			create_component(route3.$$.fragment);
    			t3 = space();
    			create_component(route4.$$.fragment);
    			t4 = space();
    			create_component(route5.$$.fragment);
    			t5 = space();
    			create_component(route6.$$.fragment);
    		},
    		l: function claim(nodes) {
    			claim_component(route0.$$.fragment, nodes);
    			t0 = claim_space(nodes);
    			claim_component(route1.$$.fragment, nodes);
    			t1 = claim_space(nodes);
    			claim_component(route2.$$.fragment, nodes);
    			t2 = claim_space(nodes);
    			claim_component(route3.$$.fragment, nodes);
    			t3 = claim_space(nodes);
    			claim_component(route4.$$.fragment, nodes);
    			t4 = claim_space(nodes);
    			claim_component(route5.$$.fragment, nodes);
    			t5 = claim_space(nodes);
    			claim_component(route6.$$.fragment, nodes);
    		},
    		m: function mount(target, anchor) {
    			mount_component(route0, target, anchor);
    			insert_hydration_dev(target, t0, anchor);
    			mount_component(route1, target, anchor);
    			insert_hydration_dev(target, t1, anchor);
    			mount_component(route2, target, anchor);
    			insert_hydration_dev(target, t2, anchor);
    			mount_component(route3, target, anchor);
    			insert_hydration_dev(target, t3, anchor);
    			mount_component(route4, target, anchor);
    			insert_hydration_dev(target, t4, anchor);
    			mount_component(route5, target, anchor);
    			insert_hydration_dev(target, t5, anchor);
    			mount_component(route6, target, anchor);
    			current = true;
    		},
    		p: noop,
    		i: function intro(local) {
    			if (current) return;
    			transition_in(route0.$$.fragment, local);
    			transition_in(route1.$$.fragment, local);
    			transition_in(route2.$$.fragment, local);
    			transition_in(route3.$$.fragment, local);
    			transition_in(route4.$$.fragment, local);
    			transition_in(route5.$$.fragment, local);
    			transition_in(route6.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(route0.$$.fragment, local);
    			transition_out(route1.$$.fragment, local);
    			transition_out(route2.$$.fragment, local);
    			transition_out(route3.$$.fragment, local);
    			transition_out(route4.$$.fragment, local);
    			transition_out(route5.$$.fragment, local);
    			transition_out(route6.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(route0, detaching);
    			if (detaching) detach_dev(t0);
    			destroy_component(route1, detaching);
    			if (detaching) detach_dev(t1);
    			destroy_component(route2, detaching);
    			if (detaching) detach_dev(t2);
    			destroy_component(route3, detaching);
    			if (detaching) detach_dev(t3);
    			destroy_component(route4, detaching);
    			if (detaching) detach_dev(t4);
    			destroy_component(route5, detaching);
    			if (detaching) detach_dev(t5);
    			destroy_component(route6, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_default_slot.name,
    		type: "slot",
    		source: "(12:0) <Router {url}>",
    		ctx
    	});

    	return block;
    }

    function create_fragment(ctx) {
    	let router;
    	let current;

    	router = new Router({
    			props: {
    				url: /*url*/ ctx[0],
    				$$slots: { default: [create_default_slot] },
    				$$scope: { ctx }
    			},
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			create_component(router.$$.fragment);
    		},
    		l: function claim(nodes) {
    			claim_component(router.$$.fragment, nodes);
    		},
    		m: function mount(target, anchor) {
    			mount_component(router, target, anchor);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			const router_changes = {};
    			if (dirty & /*url*/ 1) router_changes.url = /*url*/ ctx[0];

    			if (dirty & /*$$scope*/ 2) {
    				router_changes.$$scope = { dirty, ctx };
    			}

    			router.$set(router_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(router.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(router.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(router, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_fragment.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('App', slots, []);
    	let { url = "" } = $$props;
    	const writable_props = ['url'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<App> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('url' in $$props) $$invalidate(0, url = $$props.url);
    	};

    	$$self.$capture_state = () => ({
    		Router,
    		Route,
    		Home,
    		Generate,
    		SignIn,
    		Dashboard,
    		Delete,
    		Monitoring,
    		url
    	});

    	$$self.$inject_state = $$props => {
    		if ('url' in $$props) $$invalidate(0, url = $$props.url);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [url];
    }

    class App extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance, create_fragment, safe_not_equal, { url: 0 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "App",
    			options,
    			id: create_fragment.name
    		});
    	}

    	get url() {
    		throw new Error("<App>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set url(value) {
    		throw new Error("<App>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    new App({
        target: document.getElementById("app"),
        hydrate: true
    });

})();
//# sourceMappingURL=bundle.js.map
