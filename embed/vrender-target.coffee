tl = require 'talio'

{div, main, span, pre, nav, section, iframe, header,
 small, img, i, p, b, a, button, code,
 h1, h2, h3, h4, strong, legend, object,
 form, legend, label, input, textarea, select, label, option,
 table, thead, tbody, tfoot, tr, th, td,
 dl, dt, dd,
 ul, li} = require 'virtual-elements'

module.exports = (tab, channels) ->
  RelationshipButton = switch tab.rel
    when undefined, null, false
      (button
        'ev-click': tl.sendClick channels.openRelationshipInput, {
          target: tab.url
        }, {preventDefault: true}
      , "Establish relationship")
    else
      (form
        'ev-submit': tl.sendSubmit channels.createRelationship, {
          rel: tab.rel
          target: tab.url
        }, {preventDefault: true}
      ,
        (input
          name: 'rel'
          value: tab.rel
          'ev-input': tl.sendChange channels.changeRelationship, {
            target: tab.url
          }
        )
        (button {}, 'Save')
      )

  if tab.relExists
    (button disabled: true,
      tab.existingRel
    )
  else
    (button
      'ev-click': tl.sendClick channels.createEquality, {
        target: tab.url
      }, {preventDefault: true}
    , "Declare equal")
    (RelationshipButton)


  (div key: tab.url,
    (h1
      title: tab.url
    , tab.title)
  )
