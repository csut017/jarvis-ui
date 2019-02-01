from qi import Application
from robot import Dances, Postures, Robot
from fileDownload import SpeechGenerator

soundLocation = '/home/nao/fun/sound'
generator = SpeechGenerator(soundLocation, speechURL='http://192.168.0.8/api/speech')
generator.generateMultiple(
    greetingGeneral='Hello, it is a beautiful day.', 
    ouch='Oops, this is embarrassing', 
    timeToDance='Now, it is time to dance',
    timeToSit='I am going to sit down now')

app = Application()
with Robot('127.0.0.1', soundLocation=soundLocation) as r:
    def cancelBehaviour(event):
        r.stopCurrentBehaviour()
        r.playSound('ouch')

    r.playSound('greetingGeneral', True)
    r.registerEvent('robotHasFallen', cancelBehaviour)
    r.moveToPosture(Postures.STAND, wait=True)
    r.playSound('timeToDance', True)
    r.startBehaviour(Dances.MACARENA, wait=True)
    r.playSound('timeToSit').moveToPosture(Postures.SIT).wait()