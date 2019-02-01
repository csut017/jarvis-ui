import json
import os
import requests


class FileDownloader(object):
    def __init__(self, basePath, speechURL=None):
        self._basePath = basePath
        self._speechURL = speechURL

    def download(self, url, file):
        fullPath = os.path.join(self._basePath, file)
        print '[Downloader] Downloading from ' + url + ' to ' + fullPath
        r = requests.get(url)
        r.raise_for_status()
        with open(fullPath, 'wb') as file:
            for chunk in r.iter_content():
                file.write(chunk)

class SpeechGenerator(object):
    def __init__(self, basePath, speechURL):
        self._basePath = basePath
        self._speechURL = speechURL

    def generateMultiple(self, **kwargs):
        for file, text in kwargs.iteritems():
            self._checkFileAndDownload(file, text)

    def generate(self, file, text):
        fullPath = os.path.join(self._basePath, file) + '.wav'
        print '[Speech] Downloading speech ' + text + ' to ' + fullPath
        payload = {
            'format': 'wav',
            'text': text,
            'voice': 'female'
        }
        r = requests.post(self._speechURL, data=json.dumps(payload))
        r.raise_for_status()
        with open(fullPath, 'wb') as file:
            for chunk in r.iter_content():
                file.write(chunk)

    def _checkFileAndDownload(self, file, text):
        fullPath = os.path.join(self._basePath, file) + '.wav'
        exists = os.path.isfile(fullPath)
        if exists:
            print '[Speech] Speech ' + file + ' already exists'
        else:
            self.generate(file, text)
