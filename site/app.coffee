Promise    = require 'lie'
Router     = require 'routerjs'
tl         = require 'talio'
superagent = (require 'superagent-promise')((require 'superagent'), Promise)

# humane.js notifications
humane.timeout = 2500
humane.waitForMove = false
humane.clickToClose = true
humane.info = humane.spawn(addnCls: 'humane-flatty-info', timeout: 5000)
humane.error = humane.spawn(addnCls: 'humane-flatty-error', timeout: 4000)
humane.success = humane.spawn(addnCls: 'humane-flatty-success', timeout: 2500)

router = new Router()

State = tl.StateFactory

handlers =
  createRelationship: (State, form) ->
    Promise.resolve().then(->
      superagent
        .post('/rel/')
        .set('Content-Type', 'application/x-www-form-urlencoded')
        .send(form)
    ).then(->
      humane.success 'Relationship created!'
    ).catch((e) ->
      console.log e
      humane.error e
    )
  createEquality: (State, form) ->
    Promise.resolve().then(->
      superagent
        .post('/eql/')
        .set('Content-Type', 'application/x-www-form-urlencoded')
        .send(form)
    ).then(->
      humane.success 'Equality created!'
    ).catch((e) ->
      console.log e
      humane.error e
    )

router
  .run('#/')

tl.run document.body, (require './vrender-main'), handlers, State
