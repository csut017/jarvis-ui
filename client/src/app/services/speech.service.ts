import { Injectable } from '@angular/core';
import { environment } from '../../environments/environment'
import { LoggingService, Logger } from './logging.service';

@Injectable({
  providedIn: 'root'
})
export class SpeechService {

  constructor(logging: LoggingService) { 
    this.logger = logging.get('SpeechService');
  }

  private logger: Logger;

  say(text: string) {
    this.logger.log('Generating speech', text);
    const url = `${environment.apiURL}speech`;
    let audio = new Audio();
    audio.src = url + "?text=" + encodeURIComponent(text);
    audio.load();
    audio.play();
  }
}
