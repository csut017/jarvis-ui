import os
import qi

class Robot(object):
    ''' Defines a common interface to a Nao robot. '''

    def __init__(self, ip, soundLocation=None):
        self._ip = ip
        self._soundLocation = soundLocation
        self._session = qi.Session()
        address = 'tcp://' + ip + ':9559'
        print '[Robot] Connecting to ' + address
        self._session.connect(address)

        # Define the services that might be used
        self._audioPlayer = self._session.service('ALAudioPlayer')
        self._behavior = self._session.service('ALBehaviorManager')
        self._memory = self._session.service('ALMemory')
        self._motion = self._session.service('ALMotion')
        self._posture = self._session.service('ALRobotPosture')
        self._speech = self._session.service('ALAnimatedSpeech')
        self._system = self._session.service('ALSystem')
        self._tts = self._session.service('ALTextToSpeech')

        # Define the internal state variables
        self._currentBehaviour = None
        self._events = {}
        self._promises = []
        self._resting = True

        # Get details about the robot
        self.name = self._system.robotName()

    def _log(self, msg):
        ''' Logs a message about the robot. '''
        print '[Robot:'+ self.name + '] ' + msg

    def _prepare(self):
        ''' Prepares the robot to move. '''
        if self._resting:
            self.wakeUp()

    def __enter__(self):
        return self

    def __exit__(self, exit_type, exit_value, exit_traceback):
        self.rest()

    def moveToPosture(self, posture, wait=False):
        ''' Moves the robot to the specified posture. '''
        self._prepare()
        self._log('Moving to posture ' + posture)
        if wait:
            self._posture.goToPosture(posture, 0.8)
        else:
            p = self._posture.goToPosture(posture, 0.8, _async=True)
            self._promises.append(p)
        return self

    def playSound(self, file, wait=False):
        ''' Plays an audio file. '''
        file = file + '.wav'
        if self._soundLocation is not None:
            file = os.path.join(self._soundLocation, file)
        self._log('Playing audio ' + file)
        if wait:
            self._audioPlayer.playFile(file)
        else:
            p = self._audioPlayer.playFile(file, _async=True)
            self._promises.append(p)
        return self

    def registerEvent(self, eventName, handler):
        self._log('Registering event handler for ' + eventName)
        subscriber = self._memory.subscriber(eventName)
        subscriber.signal.connect(handler)
        self._events[eventName] = subscriber

    def rest(self, wait=False):
        self._log('Resting')
        if wait:
            self._motion.rest()
        else:
            p = self._motion.rest(_async=True)
            self._promises.append(p)
        self._resting = True
        return self

    def say(self, text, wait=False, animations=True):
        text = str(text)
        self._log('Saying ' + text)
        if wait:
            if animations:
                self._speech.say(text)
            else:
                self._tts.say(text)
        else:
            if animations:
                p = self._speech.say(text, _async=True)
            else:
                p = self._tts.say(text, _async=True)
            self._promises.append(p)
        return self

    def startBehaviour(self, behaviour, wait=False, skipCheck=False):
        ''' Starts a new behaviour. '''
        startRun = False
        if skipCheck:
            startRun = True
        else:
            if not self._currentBehaviour is None:
                if (self._behavior.isBehaviorRunning(self._currentBehaviour)):
                    raise InvalidRobotOperationError(
                            'startBehaviour', 'Another behaviour is running')
                else:
                    self._currentBehaviour = None

            self._log('Checking behaviour "' + behaviour + '" is installed')
            if (self._behavior.isBehaviorInstalled(behaviour)):
                if (not self._behavior.isBehaviorRunning(behaviour)):
                    startRun = True
                else:
                    raise InvalidRobotOperationError(
                        'startBehaviour', 'Behaviour already running')
            else:
                raise InvalidRobotOperationError(
                    'startBehaviour', 'Behaviour not found')

        if startRun:
            self._log('Starting behaviour "' + behaviour + '"')
            self._currentBehaviour = behaviour
            if wait:
                self._behavior.runBehavior(behaviour)
            else:
                p = self._behavior.runBehavior(behaviour, _async=True)
                self._promises.append(p)

        return self

    def stopCurrentBehaviour(self):
        ''' Stops the current behavour. '''
        if not self._currentBehaviour is None:
            if (self._behavior.isBehaviorRunning(self._currentBehaviour)):
                self._behavior.stopBehavior(self._currentBehaviour)
        self._currentBehaviour = None

    def wait(self):
        self._log('Waiting')
        for p in self._promises:
            self._last_result = p.value()
        self._promises = []
        return self

    def wakeUp(self):
        self._log('Waking up')
        self._motion.wakeUp()
        self._resting = False
        return self

class InvalidRobotOperationError(Exception):
    ''' An invalid operation has been requested. '''

    def __init__(self, operation, message=None):
        self._message = message
        self._operation = operation

    def __str__(self):
        return 'Cannot perform ' + self._operation + ('' if self._message is None else ': ' + self._message)

class Dances(object):
    ''' Some default dances. '''
    GANGNAM = 'gangnam-fb8eb6/gangnam'
    MACARENA = 'macarena-d73ebc/Macarena'
    TAICHI = 'taichi-7eb148/taichi'

class Postures(object):
    ''' The predefined postures for the robot. '''
    SIT = 'Sit'
    STAND = 'Stand'
    CROUCH = 'Crouch'

    LIEONBACK = 'LyingBack'
    LIEONBELLY = 'LyingBelly'
    SITBACK = 'SitRelax'
    SITFORWARD = 'Sit'
    STANDINIT = 'StandInit'
    STANDZERO = 'StandZero'
