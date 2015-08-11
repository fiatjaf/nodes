crosstab = require 'crosstab'

url = null
window.addEventListener 'message', (e) ->
  console.log 'got message from parent', e.data
  crosstab.broadcast 'message', e.data

  if e.data.kind == 'here'
    url = e.data.url
, false

crosstab.on 'message', (e) ->
  if e.origin == crosstab.id
    return

  console.log 'got message from other tab', e
  window.parent.postMessage e.data, '*'

called = false
window.onunload = window.onbeforeunload = ->
  if called
    return
  else
    called = true
    crosstab.broadcast 'message', {kind: 'nothere', url: url}
