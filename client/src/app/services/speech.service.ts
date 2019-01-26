import { Injectable } from '@angular/core';
import { environment } from '../../environments/environment'
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class SpeechService {

  constructor(logging: LoggingService) {
    this.logger = logging.get('SpeechService');
  }

  private logger: Logger;

  say(text: string): Observable<boolean> {
    const obs = Observable.create(emitter => {
      this.logger.log('Generating speech', text);
      const url = `${environment.apiURL}speech`;
      let audio = new Audio();
      audio.src = url + "?text=" + encodeURIComponent(text);
      audio.load();
      audio.onended = () => {
        this.logger.log('Speech finished');
        emitter.next(true);
      };
      audio.onerror = err => {
        this.logger.log('Unable to play speech', err);
        emitter.nexy(false);
      }
      audio.play();
    });
    return obs;
  }
}
