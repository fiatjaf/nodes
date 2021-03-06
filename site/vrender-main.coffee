tl = require 'talio'

{div, main, span, pre, nav, section,
 small, img, i, p, b, a, button, code,
 h1, h2, h3, h4, strong, legend, object,
 form, legend, label, input, textarea, select, label, option,
 table, thead, tbody, tfoot, tr, th, td,
 dl, dt, dd,
 ul, li} = require 'virtual-elements'

module.exports = (state, channels) ->
  (div {},
    (form {'ev-submit': tl.sendSubmit channels.createRelationship},
      (legend {}, 'Create a relationship')
      (input name: 'source', placeholder: 'source')
      (input name: 'rel', placeholder: 'rel')
      (input name: 'target', placeholder: 'target')
      (button {}, 'save')
    )
    (form {'ev-submit': tl.sendSubmit channels.createEquality},
      (legend {}, 'State some equality')
      (input name: 'source', placeholder: 'source')
      (input name: 'target', placeholder: 'target')
      (button {}, 'save')
    )
    (object style: {width: '100%'}, type: "image/svg+xml", data: state.graphURL) if state.graphURL
  )
