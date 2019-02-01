import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { tap, catchError } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class RoomService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('RoomService');
    this.current = new RoomInterface(http, logging, 'Test-1');
  }

  current: RoomInterface;
  private logger: Logger;
}

export class RoomInterface {
  constructor(private http: HttpClient,
    logging: LoggingService,
    public name: string) {
    this.logger = logging.get('RoomInterface');
  }

  private logger: Logger;

  getSummary(): Observable<RoomSummary> {
    const url = `${environment.apiURL}rooms/${this.name}`;
    this.logger.log('Retrieving current room summary', url);
    return this.http.get<RoomSummary>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved room summary', res)),
        catchError(this.logger.handleError<RoomSummary>(`'${this.name}'->getSummary()`))
      );
  }
}

export interface RoomSummary {
  summary: string;
}