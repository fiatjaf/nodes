tl      = require 'talio'
partial = require 'vdom-thunk'

{div, main, span, pre, nav, section, iframe, header,
 small, img, i, p, b, a, button, code,
 h1, h2, h3, h4, strong, legend, object,
 form, legend, label, input, textarea, select, label, option,
 table, thead, tbody, tfoot, tr, th, td,
 dl, dt, dd,
 ul, li} = require 'virtual-elements'

State   = require './state'
baseURL = require './baseurl'

vrenderTarget = require './vrender-target'

class IframeHook
  constructor: ->
  hook: (node) ->
    if node == window.cachedIframeNode
      return
    window.cachedIframeNode = node

    setTimeout ->
      # wire up messaging
      iframeWindow = node.contentWindow
      window.addEventListener 'message', (e) ->
        console.log 'got message from iframe', e
        switch e.data.kind
          when 'here'
            if e.data.url != State.get('here.url')
              State.change {
                otherTabs:
                  "#{e.data.url}": {title: e.data.title, url: e.data.url}
              }
          when 'left'
            State.change {
              otherTabs:
                "#{e.data.url}": false
            }
          when 'arrived'
            iframeWindow.postMessage {
              kind: 'here'
              url: (State.get 'here.url')
              title: (State.get 'here.title')
            }, '*'
      iframeWindow.postMessage {kind: 'arrived'}, '*'
      iframeWindow.postMessage {
        kind: 'here'
        url: (State.get 'here.url')
        title: (State.get 'here.title')
      }, '*'
    , 1000

module.exports = (state, channels) ->
  (div {},
    (header {},
      (h1 {}, state.here.url)
      (p {}, state.here.title)
    )
    (section {},
      (partial vrenderTarget, tab, channels) for _, tab of state.otherTabs when tab
    )
    (iframe
      src: baseURL + '/iframe.html'
      'messaging-hook': new IframeHook
      key: 'sole'
    )
  )
