
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

    function create_fragment$c(ctx) {
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
    		id: create_fragment$c.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$c($$self, $$props, $$invalidate) {
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
    		init(this, options, instance$c, create_fragment$c, safe_not_equal, { basepath: 3, url: 4 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Router",
    			options,
    			id: create_fragment$c.name
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
    function create_if_block$5(ctx) {
    	let current_block_type_index;
    	let if_block;
    	let if_block_anchor;
    	let current;
    	const if_block_creators = [create_if_block_1, create_else_block];
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
    		id: create_if_block$5.name,
    		type: "if",
    		source: "(40:0) {#if $activeRoute !== null && $activeRoute.route === route}",
    		ctx
    	});

    	return block;
    }

    // (43:2) {:else}
    function create_else_block(ctx) {
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
    		id: create_else_block.name,
    		type: "else",
    		source: "(43:2) {:else}",
    		ctx
    	});

    	return block;
    }

    // (41:2) {#if component !== null}
    function create_if_block_1(ctx) {
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
    		id: create_if_block_1.name,
    		type: "if",
    		source: "(41:2) {#if component !== null}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$b(ctx) {
    	let if_block_anchor;
    	let current;
    	let if_block = /*$activeRoute*/ ctx[1] !== null && /*$activeRoute*/ ctx[1].route === /*route*/ ctx[7] && create_if_block$5(ctx);

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
    					if_block = create_if_block$5(ctx);
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
    		id: create_fragment$b.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$b($$self, $$props, $$invalidate) {
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
    		init(this, options, instance$b, create_fragment$b, safe_not_equal, { path: 8, component: 0 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Route",
    			options,
    			id: create_fragment$b.name
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

    /* src\routes\Home.svelte generated by Svelte v3.53.1 */

    const file$9 = "src\\routes\\Home.svelte";

    function create_fragment$a(ctx) {
    	let h1;
    	let t0;
    	let t1;
    	let p;
    	let t2;

    	const block = {
    		c: function create() {
    			h1 = element("h1");
    			t0 = text("Home");
    			t1 = space();
    			p = element("p");
    			t2 = text("Welcome to my website");
    			this.h();
    		},
    		l: function claim(nodes) {
    			h1 = claim_element(nodes, "H1", { class: true });
    			var h1_nodes = children(h1);
    			t0 = claim_text(h1_nodes, "Home");
    			h1_nodes.forEach(detach_dev);
    			t1 = claim_space(nodes);
    			p = claim_element(nodes, "P", {});
    			var p_nodes = children(p);
    			t2 = claim_text(p_nodes, "Welcome to my website");
    			p_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(h1, "class", "svelte-1qbjbxu");
    			add_location(h1, file$9, 0, 0, 0);
    			add_location(p, file$9, 1, 0, 14);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, h1, anchor);
    			append_hydration_dev(h1, t0);
    			insert_hydration_dev(target, t1, anchor);
    			insert_hydration_dev(target, p, anchor);
    			append_hydration_dev(p, t2);
    		},
    		p: noop,
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(h1);
    			if (detaching) detach_dev(t1);
    			if (detaching) detach_dev(p);
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

    function instance$a($$self, $$props) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Home', slots, []);
    	const writable_props = [];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Home> was created with unknown prop '${key}'`);
    	});

    	return [];
    }

    class Home extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$a, create_fragment$a, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Home",
    			options,
    			id: create_fragment$a.name
    		});
    	}
    }

    /* src\routes\Generate.svelte generated by Svelte v3.53.1 */

    const file$8 = "src\\routes\\Generate.svelte";

    function create_fragment$9(ctx) {
    	let div7;
    	let div6;
    	let h2;
    	let t0;
    	let t1;
    	let input;
    	let t2;
    	let button0;
    	let t3;
    	let t4;
    	let button1;
    	let img;
    	let img_src_value;
    	let t5;
    	let div0;
    	let t6;
    	let t7;
    	let div2;
    	let div1;
    	let t8;
    	let div5;
    	let div3;
    	let t9;
    	let t10;
    	let div4;
    	let t11;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div7 = element("div");
    			div6 = element("div");
    			h2 = element("h2");
    			t0 = text("Generate API key");
    			t1 = space();
    			input = element("input");
    			t2 = space();
    			button0 = element("button");
    			t3 = text("Generate");
    			t4 = space();
    			button1 = element("button");
    			img = element("img");
    			t5 = space();
    			div0 = element("div");
    			t6 = text("Copied!");
    			t7 = space();
    			div2 = element("div");
    			div1 = element("div");
    			t8 = space();
    			div5 = element("div");
    			div3 = element("div");
    			t9 = text("Keep your API key safe and secure.");
    			t10 = space();
    			div4 = element("div");
    			t11 = text("API Analytics");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div7 = claim_element(nodes, "DIV", { class: true });
    			var div7_nodes = children(div7);
    			div6 = claim_element(div7_nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			h2 = claim_element(div6_nodes, "H2", { class: true });
    			var h2_nodes = children(h2);
    			t0 = claim_text(h2_nodes, "Generate API key");
    			h2_nodes.forEach(detach_dev);
    			t1 = claim_space(div6_nodes);
    			input = claim_element(div6_nodes, "INPUT", { type: true, class: true });
    			t2 = claim_space(div6_nodes);
    			button0 = claim_element(div6_nodes, "BUTTON", { id: true, class: true });
    			var button0_nodes = children(button0);
    			t3 = claim_text(button0_nodes, "Generate");
    			button0_nodes.forEach(detach_dev);
    			t4 = claim_space(div6_nodes);
    			button1 = claim_element(div6_nodes, "BUTTON", { id: true, class: true });
    			var button1_nodes = children(button1);
    			img = claim_element(button1_nodes, "IMG", { class: true, src: true, alt: true });
    			button1_nodes.forEach(detach_dev);
    			t5 = claim_space(div6_nodes);
    			div0 = claim_element(div6_nodes, "DIV", { id: true, class: true });
    			var div0_nodes = children(div0);
    			t6 = claim_text(div0_nodes, "Copied!");
    			div0_nodes.forEach(detach_dev);
    			t7 = claim_space(div6_nodes);
    			div2 = claim_element(div6_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			div1 = claim_element(div2_nodes, "DIV", { class: true, style: true });
    			children(div1).forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			t8 = claim_space(div6_nodes);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			div3 = claim_element(div5_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			t9 = claim_text(div3_nodes, "Keep your API key safe and secure.");
    			div3_nodes.forEach(detach_dev);
    			t10 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			t11 = claim_text(div4_nodes, "API Analytics");
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			div7_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(h2, "class", "svelte-x7dl2x");
    			add_location(h2, file$8, 29, 4, 847);
    			attr_dev(input, "type", "text");
    			input.readOnly = true;
    			attr_dev(input, "class", "svelte-x7dl2x");
    			add_location(input, file$8, 30, 4, 877);
    			attr_dev(button0, "id", "generateBtn");
    			attr_dev(button0, "class", "svelte-x7dl2x");
    			add_location(button0, file$8, 31, 4, 932);
    			attr_dev(img, "class", "copy-icon svelte-x7dl2x");
    			if (!src_url_equal(img.src, img_src_value = "img/copy.png")) attr_dev(img, "src", img_src_value);
    			attr_dev(img, "alt", "");
    			add_location(img, file$8, 35, 7, 1111);
    			attr_dev(button1, "id", "copyBtn");
    			attr_dev(button1, "class", "svelte-x7dl2x");
    			add_location(button1, file$8, 34, 4, 1036);
    			attr_dev(div0, "id", "copied");
    			attr_dev(div0, "class", "svelte-x7dl2x");
    			add_location(div0, file$8, 37, 4, 1181);
    			attr_dev(div1, "class", "loader");
    			set_style(div1, "display", /*loading*/ ctx[0] ? 'initial' : 'none');
    			add_location(div1, file$8, 40, 6, 1276);
    			attr_dev(div2, "class", "spinner svelte-x7dl2x");
    			add_location(div2, file$8, 39, 4, 1248);
    			attr_dev(div3, "class", "keep-secure svelte-x7dl2x");
    			add_location(div3, file$8, 44, 6, 1391);
    			attr_dev(div4, "class", "highlight logo svelte-x7dl2x");
    			add_location(div4, file$8, 45, 6, 1463);
    			attr_dev(div5, "class", "details svelte-x7dl2x");
    			add_location(div5, file$8, 43, 4, 1363);
    			attr_dev(div6, "class", "content svelte-x7dl2x");
    			add_location(div6, file$8, 28, 2, 821);
    			attr_dev(div7, "class", "generate svelte-x7dl2x");
    			add_location(div7, file$8, 27, 0, 796);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div7, anchor);
    			append_hydration_dev(div7, div6);
    			append_hydration_dev(div6, h2);
    			append_hydration_dev(h2, t0);
    			append_hydration_dev(div6, t1);
    			append_hydration_dev(div6, input);
    			set_input_value(input, /*apiKey*/ ctx[1]);
    			append_hydration_dev(div6, t2);
    			append_hydration_dev(div6, button0);
    			append_hydration_dev(button0, t3);
    			/*button0_binding*/ ctx[8](button0);
    			append_hydration_dev(div6, t4);
    			append_hydration_dev(div6, button1);
    			append_hydration_dev(button1, img);
    			/*button1_binding*/ ctx[9](button1);
    			append_hydration_dev(div6, t5);
    			append_hydration_dev(div6, div0);
    			append_hydration_dev(div0, t6);
    			/*div0_binding*/ ctx[10](div0);
    			append_hydration_dev(div6, t7);
    			append_hydration_dev(div6, div2);
    			append_hydration_dev(div2, div1);
    			append_hydration_dev(div6, t8);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, div3);
    			append_hydration_dev(div3, t9);
    			append_hydration_dev(div5, t10);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, t11);

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
    		id: create_fragment$9.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$9($$self, $$props, $$invalidate) {
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
    		init(this, options, instance$9, create_fragment$9, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Generate",
    			options,
    			id: create_fragment$9.name
    		});
    	}
    }

    /* src\routes\SignIn.svelte generated by Svelte v3.53.1 */

    const { console: console_1$2 } = globals;
    const file$7 = "src\\routes\\SignIn.svelte";

    function create_fragment$8(ctx) {
    	let div6;
    	let div5;
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
    	let div4;
    	let div2;
    	let t6;
    	let t7;
    	let div3;
    	let t8;
    	let mounted;
    	let dispose;

    	const block = {
    		c: function create() {
    			div6 = element("div");
    			div5 = element("div");
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
    			div4 = element("div");
    			div2 = element("div");
    			t6 = text("Keep your API key safe and secure.");
    			t7 = space();
    			div3 = element("div");
    			t8 = text("API Analytics");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div6 = claim_element(nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			h2 = claim_element(div5_nodes, "H2", { class: true });
    			var h2_nodes = children(h2);
    			t0 = claim_text(h2_nodes, "Dashboard");
    			h2_nodes.forEach(detach_dev);
    			t1 = claim_space(div5_nodes);

    			input = claim_element(div5_nodes, "INPUT", {
    				type: true,
    				placeholder: true,
    				class: true
    			});

    			t2 = claim_space(div5_nodes);
    			button = claim_element(div5_nodes, "BUTTON", { id: true, class: true });
    			var button_nodes = children(button);
    			t3 = claim_text(button_nodes, "Load");
    			button_nodes.forEach(detach_dev);
    			t4 = claim_space(div5_nodes);
    			div1 = claim_element(div5_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true, style: true });
    			children(div0).forEach(detach_dev);
    			div1_nodes.forEach(detach_dev);
    			t5 = claim_space(div5_nodes);
    			div4 = claim_element(div5_nodes, "DIV", { class: true });
    			var div4_nodes = children(div4);
    			div2 = claim_element(div4_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			t6 = claim_text(div2_nodes, "Keep your API key safe and secure.");
    			div2_nodes.forEach(detach_dev);
    			t7 = claim_space(div4_nodes);
    			div3 = claim_element(div4_nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			t8 = claim_text(div3_nodes, "API Analytics");
    			div3_nodes.forEach(detach_dev);
    			div4_nodes.forEach(detach_dev);
    			div5_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(h2, "class", "svelte-1nnnhku");
    			add_location(h2, file$7, 17, 4, 513);
    			attr_dev(input, "type", "text");
    			attr_dev(input, "placeholder", "Enter API key");
    			attr_dev(input, "class", "svelte-1nnnhku");
    			add_location(input, file$7, 18, 4, 536);
    			attr_dev(button, "id", "generateBtn");
    			attr_dev(button, "class", "svelte-1nnnhku");
    			add_location(button, file$7, 19, 4, 609);
    			attr_dev(div0, "class", "loader");
    			set_style(div0, "display", /*loading*/ ctx[1] ? 'initial' : 'none');
    			add_location(div0, file$7, 21, 6, 701);
    			attr_dev(div1, "class", "spinner");
    			add_location(div1, file$7, 20, 4, 673);
    			attr_dev(div2, "class", "keep-secure svelte-1nnnhku");
    			add_location(div2, file$7, 24, 6, 815);
    			attr_dev(div3, "class", "highlight logo svelte-1nnnhku");
    			add_location(div3, file$7, 25, 6, 887);
    			attr_dev(div4, "class", "details svelte-1nnnhku");
    			add_location(div4, file$7, 23, 4, 787);
    			attr_dev(div5, "class", "content svelte-1nnnhku");
    			add_location(div5, file$7, 16, 2, 487);
    			attr_dev(div6, "class", "generate svelte-1nnnhku");
    			add_location(div6, file$7, 15, 0, 462);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div6, anchor);
    			append_hydration_dev(div6, div5);
    			append_hydration_dev(div5, h2);
    			append_hydration_dev(h2, t0);
    			append_hydration_dev(div5, t1);
    			append_hydration_dev(div5, input);
    			set_input_value(input, /*apiKey*/ ctx[0]);
    			append_hydration_dev(div5, t2);
    			append_hydration_dev(div5, button);
    			append_hydration_dev(button, t3);
    			append_hydration_dev(div5, t4);
    			append_hydration_dev(div5, div1);
    			append_hydration_dev(div1, div0);
    			append_hydration_dev(div5, t5);
    			append_hydration_dev(div5, div4);
    			append_hydration_dev(div4, div2);
    			append_hydration_dev(div2, t6);
    			append_hydration_dev(div4, t7);
    			append_hydration_dev(div4, div3);
    			append_hydration_dev(div3, t8);

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
    		id: create_fragment$8.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$8($$self, $$props, $$invalidate) {
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
    		init(this, options, instance$8, create_fragment$8, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "SignIn",
    			options,
    			id: create_fragment$8.name
    		});
    	}
    }

    /* src\components\Requests.svelte generated by Svelte v3.53.1 */
    const file$6 = "src\\components\\Requests.svelte";

    // (28:2) {#if requestsPerHour != undefined}
    function create_if_block$4(ctx) {
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
    			attr_dev(div, "class", "value svelte-1bbcp2r");
    			add_location(div, file$6, 28, 4, 732);
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
    		id: create_if_block$4.name,
    		type: "if",
    		source: "(28:2) {#if requestsPerHour != undefined}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$7(ctx) {
    	let div1;
    	let div0;
    	let t0;
    	let span;
    	let t1;
    	let t2;
    	let if_block = /*requestsPerHour*/ ctx[0] != undefined && create_if_block$4(ctx);

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
    			attr_dev(span, "class", "per-hour svelte-1bbcp2r");
    			add_location(span, file$6, 25, 13, 642);
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$6, 24, 2, 603);
    			attr_dev(div1, "class", "card svelte-1bbcp2r");
    			attr_dev(div1, "title", "Last week");
    			add_location(div1, file$6, 23, 0, 563);
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
    					if_block = create_if_block$4(ctx);
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
    		id: create_fragment$7.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function thisWeek(date) {
    	let weekAgo = new Date();
    	weekAgo.setDate(weekAgo.getDate() - 7);
    	return date > weekAgo;
    }

    function instance$7($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Requests', slots, []);

    	function build() {
    		let totalRequests = 0;

    		for (let i = 0; i < data.length; i++) {
    			let date = new Date(data[i].created_at);

    			if (thisWeek(date)) {
    				totalRequests++;
    			}
    		}

    		$$invalidate(0, requestsPerHour = (24 * 7 / totalRequests).toFixed(2));
    	}

    	let requestsPerHour;

    	onMount(() => {
    		build();
    	});

    	let { data } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console.warn("<Requests> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console.warn(`<Requests> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		thisWeek,
    		build,
    		requestsPerHour,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('requestsPerHour' in $$props) $$invalidate(0, requestsPerHour = $$props.requestsPerHour);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [requestsPerHour, data];
    }

    class Requests extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$7, create_fragment$7, safe_not_equal, { data: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Requests",
    			options,
    			id: create_fragment$7.name
    		});
    	}

    	get data() {
    		throw new Error("<Requests>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Requests>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\ResponseTimes.svelte generated by Svelte v3.53.1 */
    const file$5 = "src\\components\\ResponseTimes.svelte";

    function create_fragment$6(ctx) {
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
    			t0 = text("Response Times ");
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
    			t0 = claim_text(div0_nodes, "Response Times ");
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
    			add_location(span, file$5, 60, 19, 1676);
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$5, 59, 2, 1631);
    			attr_dev(div1, "class", "value lower-quartile svelte-kx5j01");
    			add_location(div1, file$5, 63, 4, 1754);
    			attr_dev(div2, "class", "value median svelte-kx5j01");
    			add_location(div2, file$5, 64, 4, 1804);
    			attr_dev(div3, "class", "value upper-quartile svelte-kx5j01");
    			add_location(div3, file$5, 65, 4, 1850);
    			attr_dev(div4, "class", "values svelte-kx5j01");
    			add_location(div4, file$5, 62, 2, 1728);
    			attr_dev(div5, "class", "label svelte-kx5j01");
    			add_location(div5, file$5, 68, 4, 1934);
    			attr_dev(div6, "class", "label svelte-kx5j01");
    			add_location(div6, file$5, 69, 4, 1968);
    			attr_dev(div7, "class", "label svelte-kx5j01");
    			add_location(div7, file$5, 70, 4, 2005);
    			attr_dev(div8, "class", "labels svelte-kx5j01");
    			add_location(div8, file$5, 67, 2, 1908);
    			attr_dev(div9, "class", "bar-green svelte-kx5j01");
    			add_location(div9, file$5, 73, 4, 2070);
    			attr_dev(div10, "class", "bar-yellow svelte-kx5j01");
    			add_location(div10, file$5, 74, 4, 2101);
    			attr_dev(div11, "class", "bar-red svelte-kx5j01");
    			add_location(div11, file$5, 75, 4, 2133);
    			attr_dev(div12, "class", "marker svelte-kx5j01");
    			add_location(div12, file$5, 76, 4, 2162);
    			attr_dev(div13, "class", "bar svelte-kx5j01");
    			add_location(div13, file$5, 72, 2, 2047);
    			attr_dev(div14, "class", "card");
    			add_location(div14, file$5, 58, 0, 1609);
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
    			/*div12_binding*/ ctx[5](div12);
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
    			/*div12_binding*/ ctx[5](null);
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

    function instance$6($$self, $$props, $$invalidate) {
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

    		if (sorted[base + 1] !== undefined) {
    			return sorted[base] + rest * (sorted[base + 1] - sorted[base]);
    		} else {
    			return sorted[base];
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

    		$$invalidate(0, median = quantile(responseTimes, 0.25));
    		$$invalidate(1, LQ = quantile(responseTimes, 0.5));
    		$$invalidate(2, UQ = quantile(responseTimes, 0.75));
    		setMarkerPosition(median);
    	}

    	let median, LQ, UQ;
    	let marker;

    	onMount(() => {
    		build();
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
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('median' in $$props) $$invalidate(0, median = $$props.median);
    		if ('LQ' in $$props) $$invalidate(1, LQ = $$props.LQ);
    		if ('UQ' in $$props) $$invalidate(2, UQ = $$props.UQ);
    		if ('marker' in $$props) $$invalidate(3, marker = $$props.marker);
    		if ('data' in $$props) $$invalidate(4, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [median, LQ, UQ, marker, data, div12_binding];
    }

    class ResponseTimes extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$6, create_fragment$6, safe_not_equal, { data: 4 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "ResponseTimes",
    			options,
    			id: create_fragment$6.name
    		});
    	}

    	get data() {
    		throw new Error("<ResponseTimes>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<ResponseTimes>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\Endpoints.svelte generated by Svelte v3.53.1 */
    const file$4 = "src\\components\\Endpoints.svelte";

    function get_each_context$1(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[6] = list[i];
    	child_ctx[8] = i;
    	return child_ctx;
    }

    // (59:2) {#if endpoints != undefined}
    function create_if_block$3(ctx) {
    	let div;
    	let each_value = /*endpoints*/ ctx[0];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block$1(get_each_context$1(ctx, each_value, i));
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
    			attr_dev(div, "class", "endpoints svelte-1e031qh");
    			add_location(div, file$4, 59, 4, 1765);
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
    					const child_ctx = get_each_context$1(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block$1(child_ctx);
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
    		id: create_if_block$3.name,
    		type: "if",
    		source: "(59:2) {#if endpoints != undefined}",
    		ctx
    	});

    	return block;
    }

    // (61:6) {#each endpoints as endpoint, i}
    function create_each_block$1(ctx) {
    	let div6;
    	let div3;
    	let div2;
    	let div0;
    	let t0_value = /*endpoint*/ ctx[6].path + "";
    	let t0;
    	let t1;
    	let div1;
    	let t2_value = /*endpoint*/ ctx[6].count + "";
    	let t2;
    	let t3;
    	let div5;
    	let div4;
    	let t4_value = /*endpoint*/ ctx[6].path + "";
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
    			div3 = claim_element(div6_nodes, "DIV", { class: true, id: true, style: true });
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
    			attr_dev(div0, "class", "path svelte-1e031qh");
    			attr_dev(div0, "id", "endpoint-path-" + /*i*/ ctx[8]);
    			add_location(div0, file$4, 72, 14, 2262);
    			attr_dev(div1, "class", "count svelte-1e031qh");
    			attr_dev(div1, "id", "endpoint-count-" + /*i*/ ctx[8]);
    			add_location(div1, file$4, 75, 14, 2374);
    			attr_dev(div2, "class", "endpoint-label svelte-1e031qh");
    			attr_dev(div2, "id", "endpoint-label-" + /*i*/ ctx[8]);
    			add_location(div2, file$4, 71, 12, 2194);
    			attr_dev(div3, "class", "endpoint svelte-1e031qh");
    			attr_dev(div3, "id", "endpoint-" + /*i*/ ctx[8]);
    			set_style(div3, "width", /*endpoint*/ ctx[6].count / /*maxCount*/ ctx[1] * 100 + "%");

    			set_style(div3, "background", /*endpoint*/ ctx[6].status >= 200 && /*endpoint*/ ctx[6].status <= 299
    			? 'var(--highlight)'
    			: '#e46161');

    			add_location(div3, file$4, 62, 10, 1882);
    			attr_dev(div4, "class", "external-label-path");
    			add_location(div4, file$4, 79, 12, 2555);
    			attr_dev(div5, "class", "external-label svelte-1e031qh");
    			attr_dev(div5, "id", "external-label-" + /*i*/ ctx[8]);
    			add_location(div5, file$4, 78, 10, 2489);
    			attr_dev(div6, "class", "endpoint-container svelte-1e031qh");
    			add_location(div6, file$4, 61, 8, 1838);
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
    			if (dirty & /*endpoints*/ 1 && t0_value !== (t0_value = /*endpoint*/ ctx[6].path + "")) set_data_dev(t0, t0_value);
    			if (dirty & /*endpoints*/ 1 && t2_value !== (t2_value = /*endpoint*/ ctx[6].count + "")) set_data_dev(t2, t2_value);

    			if (dirty & /*endpoints, maxCount*/ 3) {
    				set_style(div3, "width", /*endpoint*/ ctx[6].count / /*maxCount*/ ctx[1] * 100 + "%");
    			}

    			if (dirty & /*endpoints*/ 1) {
    				set_style(div3, "background", /*endpoint*/ ctx[6].status >= 200 && /*endpoint*/ ctx[6].status <= 299
    				? 'var(--highlight)'
    				: '#e46161');
    			}

    			if (dirty & /*endpoints*/ 1 && t4_value !== (t4_value = /*endpoint*/ ctx[6].path + "")) set_data_dev(t4, t4_value);
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div6);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_each_block$1.name,
    		type: "each",
    		source: "(61:6) {#each endpoints as endpoint, i}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$5(ctx) {
    	let div1;
    	let div0;
    	let t0;
    	let t1;
    	let if_block = /*endpoints*/ ctx[0] != undefined && create_if_block$3(ctx);

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
    			add_location(div0, file$4, 57, 2, 1688);
    			attr_dev(div1, "class", "card");
    			add_location(div1, file$4, 56, 0, 1666);
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
    			if (detaching) detach_dev(div1);
    			if (if_block) if_block.d();
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

    function instance$5($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('Endpoints', slots, []);

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
    		endpointFreq,
    		build,
    		setEndpointLabelVisibility,
    		setEndpointLabels,
    		endpoints,
    		maxCount,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('endpoints' in $$props) $$invalidate(0, endpoints = $$props.endpoints);
    		if ('maxCount' in $$props) $$invalidate(1, maxCount = $$props.maxCount);
    		if ('data' in $$props) $$invalidate(2, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [endpoints, maxCount, data];
    }

    class Endpoints extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$5, create_fragment$5, safe_not_equal, { data: 2 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Endpoints",
    			options,
    			id: create_fragment$5.name
    		});
    	}

    	get data() {
    		throw new Error("<Endpoints>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<Endpoints>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\Footer.svelte generated by Svelte v3.53.1 */

    const file$3 = "src\\components\\Footer.svelte";

    function create_fragment$4(ctx) {
    	let div;
    	let t;

    	const block = {
    		c: function create() {
    			div = element("div");
    			t = text("API Analytics");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true });
    			var div_nodes = children(div);
    			t = claim_text(div_nodes, "API Analytics");
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "footer svelte-zlpuxh");
    			add_location(div, file$3, 0, 0, 0);
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
    		id: create_fragment$4.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function instance$4($$self, $$props) {
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
    		init(this, options, instance$4, create_fragment$4, safe_not_equal, {});

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Footer",
    			options,
    			id: create_fragment$4.name
    		});
    	}
    }

    /* src\components\SuccessRate.svelte generated by Svelte v3.53.1 */
    const file$2 = "src\\components\\SuccessRate.svelte";

    // (30:2) {#if successRate != undefined}
    function create_if_block$2(ctx) {
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
    			attr_dev(div, "class", "value svelte-yksbz4");

    			set_style(div, "color", (/*successRate*/ ctx[0] <= 75 ? 'var(--red)' : '') + (/*successRate*/ ctx[0] > 75 && /*successRate*/ ctx[0] < 90
    			? 'var(--yellow)'
    			: '') + (/*successRate*/ ctx[0] >= 90 ? 'var(--highlight)' : ''));

    			add_location(div, file$2, 30, 4, 836);
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
    		id: create_if_block$2.name,
    		type: "if",
    		source: "(30:2) {#if successRate != undefined}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$3(ctx) {
    	let div1;
    	let div0;
    	let t0;
    	let t1;
    	let if_block = /*successRate*/ ctx[0] != undefined && create_if_block$2(ctx);

    	const block = {
    		c: function create() {
    			div1 = element("div");
    			div0 = element("div");
    			t0 = text("Success Rate");
    			t1 = space();
    			if (if_block) if_block.c();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div1 = claim_element(nodes, "DIV", { class: true, title: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Success Rate");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div1_nodes);
    			if (if_block) if_block.l(div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$2, 28, 2, 754);
    			attr_dev(div1, "class", "card svelte-yksbz4");
    			attr_dev(div1, "title", "Last week");
    			add_location(div1, file$2, 27, 0, 714);
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
    					if_block = create_if_block$2(ctx);
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
    		id: create_fragment$3.name,
    		type: "component",
    		source: "",
    		ctx
    	});

    	return block;
    }

    function pastWeek(date) {
    	let weekAgo = new Date();
    	weekAgo.setDate(weekAgo.getDate() - 7);
    	return date > weekAgo;
    }

    function instance$3($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('SuccessRate', slots, []);

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

    		$$invalidate(0, successRate = successfulRequests / totalRequests * 100);
    	}

    	let successRate;

    	onMount(() => {
    		build();
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
    		pastWeek,
    		build,
    		successRate,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('successRate' in $$props) $$invalidate(0, successRate = $$props.successRate);
    		if ('data' in $$props) $$invalidate(1, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [successRate, data];
    }

    class SuccessRate extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$3, create_fragment$3, safe_not_equal, { data: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "SuccessRate",
    			options,
    			id: create_fragment$3.name
    		});
    	}

    	get data() {
    		throw new Error("<SuccessRate>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<SuccessRate>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\components\PastMonth.svelte generated by Svelte v3.53.1 */

    const { console: console_1$1 } = globals;
    const file$1 = "src\\components\\PastMonth.svelte";

    function get_each_context(ctx, list, i) {
    	const child_ctx = ctx.slice();
    	child_ctx[15] = list[i];
    	child_ctx[17] = i;
    	return child_ctx;
    }

    // (240:4) {#if successRate != undefined}
    function create_if_block$1(ctx) {
    	let div0;
    	let t0;
    	let t1;
    	let div1;
    	let each_value = /*successRate*/ ctx[0];
    	validate_each_argument(each_value);
    	let each_blocks = [];

    	for (let i = 0; i < each_value.length; i += 1) {
    		each_blocks[i] = create_each_block(get_each_context(ctx, each_value, i));
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
    			attr_dev(div0, "class", "success-rate-title");
    			add_location(div0, file$1, 240, 6, 6975);
    			attr_dev(div1, "class", "errors svelte-ik169t");
    			add_location(div1, file$1, 241, 6, 7033);
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
    			if (dirty & /*colors, Math, successRate*/ 9) {
    				each_value = /*successRate*/ ctx[0];
    				validate_each_argument(each_value);
    				let i;

    				for (i = 0; i < each_value.length; i += 1) {
    					const child_ctx = get_each_context(ctx, each_value, i);

    					if (each_blocks[i]) {
    						each_blocks[i].p(child_ctx, dirty);
    					} else {
    						each_blocks[i] = create_each_block(child_ctx);
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
    		id: create_if_block$1.name,
    		type: "if",
    		source: "(240:4) {#if successRate != undefined}",
    		ctx
    	});

    	return block;
    }

    // (243:8) {#each successRate as value, i}
    function create_each_block(ctx) {
    	let div;
    	let div_title_value;

    	const block = {
    		c: function create() {
    			div = element("div");
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", { class: true, style: true, title: true });
    			children(div).forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div, "class", "error svelte-ik169t");
    			set_style(div, "background", /*colors*/ ctx[3][Math.floor(/*value*/ ctx[15] * 10) + 1]);
    			attr_dev(div, "title", div_title_value = "" + ((/*value*/ ctx[15] * 100).toFixed(1) + "%"));
    			add_location(div, file$1, 243, 10, 7106);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    		},
    		p: function update(ctx, dirty) {
    			if (dirty & /*successRate*/ 1) {
    				set_style(div, "background", /*colors*/ ctx[3][Math.floor(/*value*/ ctx[15] * 10) + 1]);
    			}

    			if (dirty & /*successRate*/ 1 && div_title_value !== (div_title_value = "" + ((/*value*/ ctx[15] * 100).toFixed(1) + "%"))) {
    				attr_dev(div, "title", div_title_value);
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
    		source: "(243:8) {#each successRate as value, i}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$2(ctx) {
    	let div6;
    	let div0;
    	let t0;
    	let t1;
    	let div2;
    	let div1;
    	let t2;
    	let div4;
    	let div3;
    	let t3;
    	let div5;
    	let if_block = /*successRate*/ ctx[0] != undefined && create_if_block$1(ctx);

    	const block = {
    		c: function create() {
    			div6 = element("div");
    			div0 = element("div");
    			t0 = text("Past Month");
    			t1 = space();
    			div2 = element("div");
    			div1 = element("div");
    			t2 = space();
    			div4 = element("div");
    			div3 = element("div");
    			t3 = space();
    			div5 = element("div");
    			if (if_block) if_block.c();
    			this.h();
    		},
    		l: function claim(nodes) {
    			div6 = claim_element(nodes, "DIV", { class: true });
    			var div6_nodes = children(div6);
    			div0 = claim_element(div6_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			t0 = claim_text(div0_nodes, "Past Month");
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div6_nodes);
    			div2 = claim_element(div6_nodes, "DIV", { id: true });
    			var div2_nodes = children(div2);
    			div1 = claim_element(div2_nodes, "DIV", { id: true });
    			var div1_nodes = children(div1);
    			div1_nodes.forEach(detach_dev);
    			div2_nodes.forEach(detach_dev);
    			t2 = claim_space(div6_nodes);
    			div4 = claim_element(div6_nodes, "DIV", { id: true });
    			var div4_nodes = children(div4);
    			div3 = claim_element(div4_nodes, "DIV", { id: true });
    			var div3_nodes = children(div3);
    			div3_nodes.forEach(detach_dev);
    			div4_nodes.forEach(detach_dev);
    			t3 = claim_space(div6_nodes);
    			div5 = claim_element(div6_nodes, "DIV", { class: true });
    			var div5_nodes = children(div5);
    			if (if_block) if_block.l(div5_nodes);
    			div5_nodes.forEach(detach_dev);
    			div6_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "card-title");
    			add_location(div0, file$1, 227, 2, 6510);
    			attr_dev(div1, "id", "requestsFreqPlotDiv");
    			add_location(div1, file$1, 229, 4, 6577);
    			attr_dev(div2, "id", "plotly");
    			add_location(div2, file$1, 228, 2, 6554);
    			attr_dev(div3, "id", "responseTimePlotDiv");
    			add_location(div3, file$1, 234, 4, 6748);
    			attr_dev(div4, "id", "plotlyy");
    			add_location(div4, file$1, 233, 2, 6724);
    			attr_dev(div5, "class", "success-rate-container svelte-ik169t");
    			add_location(div5, file$1, 238, 2, 6895);
    			attr_dev(div6, "class", "card svelte-ik169t");
    			add_location(div6, file$1, 226, 0, 6488);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div6, anchor);
    			append_hydration_dev(div6, div0);
    			append_hydration_dev(div0, t0);
    			append_hydration_dev(div6, t1);
    			append_hydration_dev(div6, div2);
    			append_hydration_dev(div2, div1);
    			/*div1_binding*/ ctx[5](div1);
    			append_hydration_dev(div6, t2);
    			append_hydration_dev(div6, div4);
    			append_hydration_dev(div4, div3);
    			/*div3_binding*/ ctx[6](div3);
    			append_hydration_dev(div6, t3);
    			append_hydration_dev(div6, div5);
    			if (if_block) if_block.m(div5, null);
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*successRate*/ ctx[0] != undefined) {
    				if (if_block) {
    					if_block.p(ctx, dirty);
    				} else {
    					if_block = create_if_block$1(ctx);
    					if_block.c();
    					if_block.m(div5, null);
    				}
    			} else if (if_block) {
    				if_block.d(1);
    				if_block = null;
    			}
    		},
    		i: noop,
    		o: noop,
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div6);
    			/*div1_binding*/ ctx[5](null);
    			/*div3_binding*/ ctx[6](null);
    			if (if_block) if_block.d();
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
    			// showline: false,
    			// zeroline: false,
    			fixedrange: true
    		}, // visible: false,
    		xaxis: {
    			title: { text: "Date" },
    			// linecolor: "black",
    			showgrid: false,
    			// showline: false,
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
    			// showline: false,
    			// zeroline: false,
    			fixedrange: true
    		},
    		xaxis: {
    			title: { text: "Date" },
    			// linecolor: "black",
    			// showgrid: false,
    			// showline: false,
    			fixedrange: true,
    			range: [monthAgo, tomorrow],
    			visible: false
    		},
    		dragmode: false
    	};
    }

    function instance$2($$self, $$props, $$invalidate) {
    	let { $$slots: slots = {}, $$scope } = $$props;
    	validate_slots('PastMonth', slots, []);

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

    		console.log(successArr);
    		$$invalidate(0, successRate = successArr);
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
    		console.log(plotData);

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
    		console.log(plotData);

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

    	$$self.$$.on_mount.push(function () {
    		if (data === undefined && !('data' in $$props || $$self.$$.bound[$$self.$$.props['data']])) {
    			console_1$1.warn("<PastMonth> was created without expected prop 'data'");
    		}
    	});

    	const writable_props = ['data'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console_1$1.warn(`<PastMonth> was created with unknown prop '${key}'`);
    	});

    	function div1_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			requestsFreqPlotDiv = $$value;
    			$$invalidate(2, requestsFreqPlotDiv);
    		});
    	}

    	function div3_binding($$value) {
    		binding_callbacks[$$value ? 'unshift' : 'push'](() => {
    			responseTimePlotDiv = $$value;
    			$$invalidate(1, responseTimePlotDiv);
    		});
    	}

    	$$self.$$set = $$props => {
    		if ('data' in $$props) $$invalidate(4, data = $$props.data);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		pastMonth,
    		colors,
    		daysAgo,
    		setSuccessRate,
    		responseTimePlotLayout,
    		responseTimeLine,
    		responseTimePlotData,
    		genResponseTimePlot,
    		requestsFreqPlotLayout,
    		requestsFreqLine,
    		requestsFreqPlotData,
    		genRequestsFreqPlot,
    		build,
    		successRate,
    		responseTimePlotDiv,
    		requestsFreqPlotDiv,
    		data
    	});

    	$$self.$inject_state = $$props => {
    		if ('colors' in $$props) $$invalidate(3, colors = $$props.colors);
    		if ('successRate' in $$props) $$invalidate(0, successRate = $$props.successRate);
    		if ('responseTimePlotDiv' in $$props) $$invalidate(1, responseTimePlotDiv = $$props.responseTimePlotDiv);
    		if ('requestsFreqPlotDiv' in $$props) $$invalidate(2, requestsFreqPlotDiv = $$props.requestsFreqPlotDiv);
    		if ('data' in $$props) $$invalidate(4, data = $$props.data);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [
    		successRate,
    		responseTimePlotDiv,
    		requestsFreqPlotDiv,
    		colors,
    		data,
    		div1_binding,
    		div3_binding
    	];
    }

    class PastMonth extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$2, create_fragment$2, safe_not_equal, { data: 4 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "PastMonth",
    			options,
    			id: create_fragment$2.name
    		});
    	}

    	get data() {
    		throw new Error("<PastMonth>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set data(value) {
    		throw new Error("<PastMonth>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\routes\Dashboard.svelte generated by Svelte v3.53.1 */

    const { console: console_1 } = globals;
    const file = "src\\routes\\Dashboard.svelte";

    // (30:2) {#if data != undefined}
    function create_if_block(ctx) {
    	let div3;
    	let div1;
    	let div0;
    	let requests;
    	let t0;
    	let successrate;
    	let t1;
    	let responsetimes;
    	let t2;
    	let endpoints;
    	let t3;
    	let div2;
    	let pastmonth;
    	let current;

    	requests = new Requests({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	successrate = new SuccessRate({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	responsetimes = new ResponseTimes({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	endpoints = new Endpoints({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	pastmonth = new PastMonth({
    			props: { data: /*data*/ ctx[0] },
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			div3 = element("div");
    			div1 = element("div");
    			div0 = element("div");
    			create_component(requests.$$.fragment);
    			t0 = space();
    			create_component(successrate.$$.fragment);
    			t1 = space();
    			create_component(responsetimes.$$.fragment);
    			t2 = space();
    			create_component(endpoints.$$.fragment);
    			t3 = space();
    			div2 = element("div");
    			create_component(pastmonth.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div3 = claim_element(nodes, "DIV", { class: true });
    			var div3_nodes = children(div3);
    			div1 = claim_element(div3_nodes, "DIV", { class: true });
    			var div1_nodes = children(div1);
    			div0 = claim_element(div1_nodes, "DIV", { class: true });
    			var div0_nodes = children(div0);
    			claim_component(requests.$$.fragment, div0_nodes);
    			t0 = claim_space(div0_nodes);
    			claim_component(successrate.$$.fragment, div0_nodes);
    			div0_nodes.forEach(detach_dev);
    			t1 = claim_space(div1_nodes);
    			claim_component(responsetimes.$$.fragment, div1_nodes);
    			t2 = claim_space(div1_nodes);
    			claim_component(endpoints.$$.fragment, div1_nodes);
    			div1_nodes.forEach(detach_dev);
    			t3 = claim_space(div3_nodes);
    			div2 = claim_element(div3_nodes, "DIV", { class: true });
    			var div2_nodes = children(div2);
    			claim_component(pastmonth.$$.fragment, div2_nodes);
    			div2_nodes.forEach(detach_dev);
    			div3_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			attr_dev(div0, "class", "two-row svelte-1n7lknr");
    			add_location(div0, file, 32, 8, 1102);
    			attr_dev(div1, "class", "left");
    			add_location(div1, file, 31, 6, 1074);
    			attr_dev(div2, "class", "right svelte-1n7lknr");
    			add_location(div2, file, 39, 6, 1290);
    			attr_dev(div3, "class", "dashboard svelte-1n7lknr");
    			add_location(div3, file, 30, 4, 1043);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div3, anchor);
    			append_hydration_dev(div3, div1);
    			append_hydration_dev(div1, div0);
    			mount_component(requests, div0, null);
    			append_hydration_dev(div0, t0);
    			mount_component(successrate, div0, null);
    			append_hydration_dev(div1, t1);
    			mount_component(responsetimes, div1, null);
    			append_hydration_dev(div1, t2);
    			mount_component(endpoints, div1, null);
    			append_hydration_dev(div3, t3);
    			append_hydration_dev(div3, div2);
    			mount_component(pastmonth, div2, null);
    			current = true;
    		},
    		p: function update(ctx, dirty) {
    			const requests_changes = {};
    			if (dirty & /*data*/ 1) requests_changes.data = /*data*/ ctx[0];
    			requests.$set(requests_changes);
    			const successrate_changes = {};
    			if (dirty & /*data*/ 1) successrate_changes.data = /*data*/ ctx[0];
    			successrate.$set(successrate_changes);
    			const responsetimes_changes = {};
    			if (dirty & /*data*/ 1) responsetimes_changes.data = /*data*/ ctx[0];
    			responsetimes.$set(responsetimes_changes);
    			const endpoints_changes = {};
    			if (dirty & /*data*/ 1) endpoints_changes.data = /*data*/ ctx[0];
    			endpoints.$set(endpoints_changes);
    			const pastmonth_changes = {};
    			if (dirty & /*data*/ 1) pastmonth_changes.data = /*data*/ ctx[0];
    			pastmonth.$set(pastmonth_changes);
    		},
    		i: function intro(local) {
    			if (current) return;
    			transition_in(requests.$$.fragment, local);
    			transition_in(successrate.$$.fragment, local);
    			transition_in(responsetimes.$$.fragment, local);
    			transition_in(endpoints.$$.fragment, local);
    			transition_in(pastmonth.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(requests.$$.fragment, local);
    			transition_out(successrate.$$.fragment, local);
    			transition_out(responsetimes.$$.fragment, local);
    			transition_out(endpoints.$$.fragment, local);
    			transition_out(pastmonth.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div3);
    			destroy_component(requests);
    			destroy_component(successrate);
    			destroy_component(responsetimes);
    			destroy_component(endpoints);
    			destroy_component(pastmonth);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_if_block.name,
    		type: "if",
    		source: "(30:2) {#if data != undefined}",
    		ctx
    	});

    	return block;
    }

    function create_fragment$1(ctx) {
    	let div;
    	let t;
    	let footer;
    	let current;
    	let if_block = /*data*/ ctx[0] != undefined && create_if_block(ctx);
    	footer = new Footer({ $$inline: true });

    	const block = {
    		c: function create() {
    			div = element("div");
    			if (if_block) if_block.c();
    			t = space();
    			create_component(footer.$$.fragment);
    			this.h();
    		},
    		l: function claim(nodes) {
    			div = claim_element(nodes, "DIV", {});
    			var div_nodes = children(div);
    			if (if_block) if_block.l(div_nodes);
    			t = claim_space(div_nodes);
    			claim_component(footer.$$.fragment, div_nodes);
    			div_nodes.forEach(detach_dev);
    			this.h();
    		},
    		h: function hydrate() {
    			add_location(div, file, 27, 0, 974);
    		},
    		m: function mount(target, anchor) {
    			insert_hydration_dev(target, div, anchor);
    			if (if_block) if_block.m(div, null);
    			append_hydration_dev(div, t);
    			mount_component(footer, div, null);
    			current = true;
    		},
    		p: function update(ctx, [dirty]) {
    			if (/*data*/ ctx[0] != undefined) {
    				if (if_block) {
    					if_block.p(ctx, dirty);

    					if (dirty & /*data*/ 1) {
    						transition_in(if_block, 1);
    					}
    				} else {
    					if_block = create_if_block(ctx);
    					if_block.c();
    					transition_in(if_block, 1);
    					if_block.m(div, t);
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
    			transition_in(footer.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(if_block);
    			transition_out(footer.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			if (detaching) detach_dev(div);
    			if (if_block) if_block.d();
    			destroy_component(footer);
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
    	validate_slots('Dashboard', slots, []);

    	async function fetchData() {
    		$$invalidate(1, userID = formatUUID(userID));

    		// Fetch page ID
    		const response = await fetch(`https://api-analytics-server.vercel.app/api/data/${userID}`);

    		if (response.status == 200) {
    			const json = await response.json();
    			$$invalidate(0, data = json.value);
    			console.log(data);
    		}
    	}

    	let data;

    	onMount(() => {
    		fetchData();
    	});

    	let { userID } = $$props;

    	$$self.$$.on_mount.push(function () {
    		if (userID === undefined && !('userID' in $$props || $$self.$$.bound[$$self.$$.props['userID']])) {
    			console_1.warn("<Dashboard> was created without expected prop 'userID'");
    		}
    	});

    	const writable_props = ['userID'];

    	Object.keys($$props).forEach(key => {
    		if (!~writable_props.indexOf(key) && key.slice(0, 2) !== '$$' && key !== 'slot') console_1.warn(`<Dashboard> was created with unknown prop '${key}'`);
    	});

    	$$self.$$set = $$props => {
    		if ('userID' in $$props) $$invalidate(1, userID = $$props.userID);
    	};

    	$$self.$capture_state = () => ({
    		onMount,
    		Requests,
    		ResponseTimes,
    		Endpoints,
    		Footer,
    		SuccessRate,
    		PastMonth,
    		formatUUID,
    		fetchData,
    		data,
    		userID
    	});

    	$$self.$inject_state = $$props => {
    		if ('data' in $$props) $$invalidate(0, data = $$props.data);
    		if ('userID' in $$props) $$invalidate(1, userID = $$props.userID);
    	};

    	if ($$props && "$$inject" in $$props) {
    		$$self.$inject_state($$props.$$inject);
    	}

    	return [data, userID];
    }

    class Dashboard extends SvelteComponentDev {
    	constructor(options) {
    		super(options);
    		init(this, options, instance$1, create_fragment$1, safe_not_equal, { userID: 1 });

    		dispatch_dev("SvelteRegisterComponent", {
    			component: this,
    			tagName: "Dashboard",
    			options,
    			id: create_fragment$1.name
    		});
    	}

    	get userID() {
    		throw new Error("<Dashboard>: Props cannot be read directly from the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}

    	set userID(value) {
    		throw new Error("<Dashboard>: Props cannot be set directly on the component instance unless compiling with 'accessors: true' or '<svelte:options accessors/>'");
    	}
    }

    /* src\App.svelte generated by Svelte v3.53.1 */

    // (10:0) <Router {url}>
    function create_default_slot(ctx) {
    	let route0;
    	let t0;
    	let route1;
    	let t1;
    	let route2;
    	let current;

    	route0 = new Route({
    			props: { path: "/generate", component: Generate },
    			$$inline: true
    		});

    	route1 = new Route({
    			props: { path: "/dashboard", component: SignIn },
    			$$inline: true
    		});

    	route2 = new Route({
    			props: {
    				path: "/dashboard/:userID",
    				component: Dashboard
    			},
    			$$inline: true
    		});

    	const block = {
    		c: function create() {
    			create_component(route0.$$.fragment);
    			t0 = space();
    			create_component(route1.$$.fragment);
    			t1 = space();
    			create_component(route2.$$.fragment);
    		},
    		l: function claim(nodes) {
    			claim_component(route0.$$.fragment, nodes);
    			t0 = claim_space(nodes);
    			claim_component(route1.$$.fragment, nodes);
    			t1 = claim_space(nodes);
    			claim_component(route2.$$.fragment, nodes);
    		},
    		m: function mount(target, anchor) {
    			mount_component(route0, target, anchor);
    			insert_hydration_dev(target, t0, anchor);
    			mount_component(route1, target, anchor);
    			insert_hydration_dev(target, t1, anchor);
    			mount_component(route2, target, anchor);
    			current = true;
    		},
    		p: noop,
    		i: function intro(local) {
    			if (current) return;
    			transition_in(route0.$$.fragment, local);
    			transition_in(route1.$$.fragment, local);
    			transition_in(route2.$$.fragment, local);
    			current = true;
    		},
    		o: function outro(local) {
    			transition_out(route0.$$.fragment, local);
    			transition_out(route1.$$.fragment, local);
    			transition_out(route2.$$.fragment, local);
    			current = false;
    		},
    		d: function destroy(detaching) {
    			destroy_component(route0, detaching);
    			if (detaching) detach_dev(t0);
    			destroy_component(route1, detaching);
    			if (detaching) detach_dev(t1);
    			destroy_component(route2, detaching);
    		}
    	};

    	dispatch_dev("SvelteRegisterBlock", {
    		block,
    		id: create_default_slot.name,
    		type: "slot",
    		source: "(10:0) <Router {url}>",
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
